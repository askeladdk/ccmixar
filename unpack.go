package main

import (
	"io"

	"golang.org/x/crypto/blowfish"
)

type mixFileEntry struct {
	id     uint32
	offset uint32
	size   uint32
	name   string
}

type mixFileEntries []mixFileEntry

func (xs mixFileEntries) indexByID(id uint32) int {
	for i := 0; i < len(xs); i++ {
		if xs[i].id == id {
			return i
		}
	}
	return -1
}

type mixFile struct {
	files  mixFileEntries
	flags  uint32
	size   uint32
	offset uint32
	reader *io.SectionReader
	game   gameID
	keysrc []byte
}

func readFileEntries(r io.Reader, count uint16) (mixFileEntries, error) {
	var entries mixFileEntries
	for i := uint16(0); i < count; i++ {
		if id, err := readUint32(r); err != nil {
			return nil, err
		} else if offset, err := readUint32(r); err != nil {
			return nil, err
		} else if size, err := readUint32(r); err != nil {
			return nil, err
		} else {
			entries = append(entries, mixFileEntry{
				id:     id,
				offset: offset,
				size:   size,
			})
		}
	}
	return entries, nil
}

func readIndex(r io.Reader, count uint16) (uint32, []mixFileEntry, error) {
	if size, err := readUint32(r); err != nil {
		return 0, nil, err
	} else if files, err := readFileEntries(r, count); err != nil {
		return 0, nil, err
	} else {
		return size, files, nil
	}
}

func unpackMixFile(r *io.SectionReader, game gameID) (*mixFile, error) {
	if count, err := readUint16(r); err != nil {
		return nil, err
	} else if count != 0 {
		_, files, err := readIndex(r, count)
		if err != nil {
			return nil, err
		}
		offset := 6 + 12*uint32(count)
		return &mixFile{
			files:  files,
			flags:  0,
			size:   uint32(r.Size()) - offset,
			offset: offset,
			reader: r,
			game:   gameCC1,
		}, nil
	} else if flags16, err := readUint16(r); err != nil {
		return nil, err
	} else {
		flags := uint32(flags16) << 16

		if (flags & flagEncrypted) != 0 {
			keySource := [80]byte{}
			_, err := r.Read(keySource[:])
			if err != nil {
				return nil, err
			}
			blowfishKey := blowfishKeyFromKeySource(keySource[:])
			cipher, err := blowfish.NewCipher(blowfishKey)
			if err != nil {
				return nil, err
			}
			ecb := newECBReader(r, cipher)
			if count, err := readUint16(ecb); err != nil {
				return nil, err
			} else if _, files, err := readIndex(ecb, count); err != nil {
				return nil, err
			} else {
				offset := 84 + ((6 + 12*uint32(count) + 7) &^ 7)
				return &mixFile{
					files:  files,
					flags:  flags,
					size:   uint32(r.Size()) - offset,
					offset: offset,
					reader: r,
					game:   game,
					keysrc: keySource[:],
				}, nil
			}
		} else if count, err := readUint16(r); err != nil {
			return nil, err
		} else if _, files, err := readIndex(r, count); err != nil {
			return nil, err
		} else {
			offset := 10 + 12*uint32(count)
			return &mixFile{
				files:  files,
				flags:  flags,
				size:   uint32(r.Size()) - offset,
				offset: offset,
				reader: r,
				game:   game,
			}, nil
		}
	}
}

func (mix *mixFile) OpenFile(i int) *io.SectionReader {
	info := mix.files[i]
	return io.NewSectionReader(mix.reader, int64(mix.offset+info.offset), int64(info.size))
}

func (mix *mixFile) ReadLmd() error {
	lmdID := getLmdFileID(mix.game)
	if fileIndex := mix.files.indexByID(lmdID); fileIndex == -1 {
		return nil
	} else if mapper, err := lmdRead(mix.OpenFile(fileIndex)); err != nil {
		return err
	} else {
		for i := 0; i < len(mix.files); i++ {
			if name, ok := mapper[mix.files[i].id]; ok {
				mix.files[i].name = name
			}
		}
		return nil
	}
}
