package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var defaultKeySource = []byte{
	0xca, 0xd0, 0xb0, 0x1b, 0xfe, 0x3f, 0x3f, 0xb6,
	0xca, 0xc0, 0xbd, 0x8f, 0x40, 0xf0, 0xee, 0x85,
	0x6e, 0xe1, 0xda, 0x7a, 0xef, 0xb4, 0xd4, 0xbb,
	0x6a, 0xd8, 0x4b, 0x84, 0x26, 0x99, 0x6f, 0xfd,
	0x65, 0x97, 0xf2, 0x5f, 0xa4, 0x46, 0xdb, 0x47,
	0x88, 0x63, 0x4f, 0x2c, 0x14, 0x0b, 0x3c, 0xce,
	0xaa, 0xc4, 0x5c, 0xe4, 0x15, 0x86, 0x26, 0x5c,
	0x52, 0x3a, 0x80, 0xf8, 0xbe, 0x45, 0x40, 0x6a,
	0x66, 0xb4, 0xc5, 0xf6, 0xd0, 0x12, 0xe0, 0x43,
	0x44, 0x65, 0xc6, 0xe3, 0x9e, 0xf9, 0x43, 0x35,
}

func stringToGameID(s string) (gameID, error) {
	switch strings.ToLower(s) {
	case "cc1":
		return gameCC1, nil
	case "cc2":
		return gameCC2, nil
	case "ra1":
		return gameRA1, nil
	case "ra2":
		return gameRA2, nil
	case "":
		return 0, errors.New("no game specified")
	default:
		return 0, fmt.Errorf("invalid game: %s", s)
	}
}

func commandPack(args []string) error {
	var (
		cmd      = flag.NewFlagSet("pack", flag.ExitOnError)
		dirname  = cmd.String("dir", "", "Path to input directory.")
		filename = cmd.String("mix", "out.mix", "Path to output .mix file.")
		game     = cmd.String("game", "", "One of cc1, cc2, ra1, ra2.")
		checksum = cmd.Bool("checksum", false, "Compute checksum if game is not cc1.")
		database = cmd.Bool("database", false, "Include local mix database.")
		encrypt  = cmd.Bool("encrypt", false, "Encrypt if game is not cc1.")
	)

	if err := cmd.Parse(args); err != nil {
		return err
	}

	if *dirname == "" {
		return errors.New("no directory specified")
	} else if *filename == "" {
		return errors.New("no output file specified")
	}

	absdirname, _ := filepath.Abs(*dirname)
	absfilename, _ := filepath.Abs(*filename)
	if filepath.Dir(absfilename) == absdirname {
		return errors.New("cannot output to the directory that is being packed")
	}

	flags := uint32(0)
	if *checksum {
		flags |= flagChecksum
	}
	if *encrypt {
		flags |= flagEncrypted
	}

	if gameID, err := stringToGameID(*game); err != nil {
		return err
	} else if f, err := os.OpenFile(absfilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644); err != nil {
		return err
	} else {
		defer f.Close()
		files, err := listFilesToPack(absdirname, *database, gameID)
		if err != nil {
			return err
		}
		wb := bufio.NewWriter(f)
		if err := pack(wb, files, gameID, flags, defaultKeySource); err != nil {
			return err
		} else if err := wb.Flush(); err != nil {
			return err
		}
	}

	return nil
}

func commandUnpack(args []string) error {
	var (
		cmd      = flag.NewFlagSet("unpack", flag.ExitOnError)
		filename = cmd.String("mix", "", "Path to .mix file.")
		dirname  = cmd.String("dir", "", "Output directory.")
		game     = cmd.String("game", "", "One of cc1, cc2, ra1, ra2.")
		gmd      = cmd.String("csv", "", "Path to mix database csv.")
	)

	if err := cmd.Parse(args); err != nil {
		return err
	} else if *dirname == "" {
		return errors.New("no output directory specified")
	} else if *filename == "" {
		return errors.New("no mix file specified")
	}

	absdirname, _ := filepath.Abs(*dirname)
	absfilename, _ := filepath.Abs(*filename)
	if filepath.Dir(absfilename) == absdirname {
		return errors.New("cannot output to the same directory where the input mix file is located")
	} else if err := os.MkdirAll(absdirname, os.ModePerm); err != nil {
		return err
	} else if gameID, err := stringToGameID(*game); err != nil {
		return err
	} else if f, err := os.Open(*filename); err != nil {
		return err
	} else if stat, err := f.Stat(); err != nil {
		return err
	} else if mix, err := unpackMixFile(io.NewSectionReader(f, 0, stat.Size()), gameID); err != nil {
		return err
	} else {
		defer f.Close()

		_ = mix.ReadGmd(*gmd)
		mix.RecoverLmd()
		_ = mix.ReadLmd()

		for i, entry := range mix.files {
			infile := mix.OpenFile(i)
			fname := entry.name
			if fname == "" {
				fname = fmt.Sprintf("%08X", entry.id)
			}
			fname = filepath.Join(absdirname, fname)
			if outfile, err := os.OpenFile(fname, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644); err != nil {
				return err
			} else if _, err := io.Copy(outfile, infile); err != nil {
				return err
			} else if err := outfile.Close(); err != nil {
				return err
			}
		}

		return nil
	}
}

func commandInfo(args []string) error {
	var (
		cmd      = flag.NewFlagSet("info", flag.ExitOnError)
		filename = cmd.String("mix", "", "Path to .mix file.")
		game     = cmd.String("game", "", "One of cc1, cc2, ra1, ra2.")
		gmd      = cmd.String("csv", "", "Path to mix database csv.")
	)

	if err := cmd.Parse(args); err != nil {
		return err
	} else if len(*filename) == 0 {
		return errors.New("no mix file specified")
	}

	if gameID, err := stringToGameID(*game); err != nil {
		return err
	} else if f, err := os.Open(*filename); err != nil {
		return err
	} else if stat, err := f.Stat(); err != nil {
		return err
	} else if mix, err := unpackMixFile(io.NewSectionReader(f, 0, stat.Size()), gameID); err != nil {
		return err
	} else {
		defer f.Close()

		_ = mix.ReadGmd(*gmd)
		mix.RecoverLmd()
		_ = mix.ReadLmd()

		fmt.Printf("checksum:  %t\n", (mix.flags&flagChecksum) != 0)
		fmt.Printf("encrypted: %t\n", (mix.flags&flagEncrypted) != 0)
		fmt.Printf("files:     %d\n", len(mix.files))
		fmt.Printf("size:      %d bytes\n", mix.size)

		for i, entry := range mix.files {
			fmt.Printf("%04d - %08X %08X % 12d %s\n", i, entry.id, entry.offset, entry.size, entry.name)
		}

		return nil
	}
}

func commandRepair(args []string) error {
	var (
		cmd      = flag.NewFlagSet("repair", flag.ExitOnError)
		filename = cmd.String("mix", "", "Path to .mix file.")
		game     = cmd.String("game", "", "One of cc1, cc2, ra1, ra2.")
	)

	if err := cmd.Parse(args); err != nil {
		return err
	} else if len(*filename) == 0 {
		return errors.New("no mix file specified")
	}

	if gameID, err := stringToGameID(*game); err != nil {
		return err
	} else if f, err := os.OpenFile(*filename, os.O_RDWR, 0); err != nil {
		return err
	} else if stat, err := f.Stat(); err != nil {
		return err
	} else if mix, err := unpackMixFile(io.NewSectionReader(f, 0, stat.Size()), gameID); err != nil {
		return err
	} else {
		defer f.Close()

		mix.RecoverLmd()
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return err
		}

		if err := mix.RewriteHeader(f); err != nil {
			return err
		}

		return nil
	}
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: ccmixar <command> [<args>]")
		fmt.Println("  command:")
		fmt.Println("    info   Lists mix file contents.")
		fmt.Println("    pack   Packs a directory in a mix file.")
		fmt.Println("    repair Repairs a mangled mix file.")
		fmt.Println("    unpack Unpacks a mix file to a directory.")
		return
	}

	var cmderr error
	switch os.Args[1] {
	case "pack":
		cmderr = commandPack(os.Args[2:])
	case "info":
		cmderr = commandInfo(os.Args[2:])
	case "unpack":
		cmderr = commandUnpack(os.Args[2:])
	case "repair":
		cmderr = commandRepair(os.Args[2:])
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}

	if cmderr != nil {
		fmt.Println(cmderr.Error())
		os.Exit(1)
	}
}
