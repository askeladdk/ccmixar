package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
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
		return errors.New("No directory specified.")
	} else if *filename == "" {
		return errors.New("No output file specified.")
	}

	absdirname, _ := filepath.Abs(*dirname)
	absfilename, _ := filepath.Abs(*filename)
	if filepath.Dir(absfilename) == absdirname {
		return errors.New("Cannot output to the directory that is being packed.")
	}

	var gameId gameId

	switch strings.ToLower(*game) {
	case "cc1":
		gameId = gameId_CC1
	case "cc2":
		gameId = gameId_CC2
	case "ra1":
		gameId = gameId_RA1
	case "ra2":
		gameId = gameId_RA2
	case "":
		return errors.New("No game specified.")
	default:
		return errors.New(fmt.Sprintf("Invalid game: %s.", *game))
	}

	flags := uint32(0)
	if *checksum {
		flags |= flagChecksum
	}
	if *encrypt {
		flags |= flagEncrypted
	}

	if f, err := os.OpenFile(absfilename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644); err != nil {
		return err
	} else {
		defer f.Close()
		if files, err := listFilesToPack(absdirname, *database, gameId); err != nil {
			return err
		} else {
			wb := bufio.NewWriter(f)
			if err := pack(wb, files, gameId, flags, defaultKeySource); err != nil {
				return err
			} else if err := wb.Flush(); err != nil {
				return err
			}
		}
	}

	return nil
}

func commandInfo(args []string) error {
	var (
		cmd      = flag.NewFlagSet("info", flag.ExitOnError)
		filename = cmd.String("mix", "", "Path to .mix file.")
	)

	if err := cmd.Parse(args); err != nil {
		return err
	} else if len(*filename) == 0 {
		return errors.New("No mix file specified.")
	}

	if f, err := os.Open(*filename); err != nil {
		return err
	} else if mix, err := readMixFile(f); err != nil {
		return err
	} else {
		var flags []string
		if (mix.flags & flagChecksum) != 0 {
			flags = append(flags, "checksum")
		}
		if (mix.flags & flagEncrypted) != 0 {
			flags = append(flags, "encrypted")
		}

		fmt.Printf("size: %d bytes\n", mix.size)
		if len(flags) != 0 {
			fmt.Printf("flags: %s\n", strings.Join(flags, ", "))
		}
		fmt.Printf("file     offset\n")
		for _, entry := range mix.files {
			fmt.Printf("%08X %08X % 10d bytes\n", entry.id, entry.offset, entry.size)
		}
	}

	return nil
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: ccmixar <command> [<args>]")
		fmt.Println("  command:")
		fmt.Println("    pack Packs a directory in a mix file.")
		fmt.Println("    info Lists mix file contents.")
		return
	}

	var cmderr error
	switch os.Args[1] {
	case "pack":
		cmderr = commandPack(os.Args[2:])
	case "info":
		cmderr = commandInfo(os.Args[2:])
	default:
		fmt.Printf("%q is not valid command.\n", os.Args[1])
		os.Exit(2)
	}

	if cmderr != nil {
		fmt.Println(cmderr.Error())
		os.Exit(1)
	}
}
