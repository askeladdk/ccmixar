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

func (this mixFileEntries) indexById(id uint32) int {
	for i := 0; i < len(this); i++ {
		if this[i].id == id {
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

func unpackMixFile(r *io.SectionReader) (*mixFile, error) {
	if count, err := readUint16(r); err != nil {
		return nil, err
	} else if count != 0 {
		if size, files, err := readIndex(r, count); err != nil {
			return nil, err
		} else {
			return &mixFile{
				files:  files,
				flags:  0,
				size:   size,
				offset: 6 + 12*uint32(count),
				reader: r,
			}, nil
		}
	} else if flags16, err := readUint16(r); err != nil {
		return nil, err
	} else {
		flags := uint32(flags16) << 16

		if (flags & flagEncrypted) != 0 {
			keySource := [80]byte{}
			if _, err := r.Read(keySource[:]); err != nil {
				return nil, err
			} else {
				blowfishKey := blowfishKeyFromKeySource(keySource[:])
				if cipher, err := blowfish.NewCipher(blowfishKey); err != nil {
					return nil, err
				} else {
					ecb := newECBReader(r, cipher)
					if count, err := readUint16(ecb); err != nil {
						return nil, err
					} else if size, files, err := readIndex(ecb, count); err != nil {
						return nil, err
					} else {
						return &mixFile{
							files:  files,
							flags:  flags,
							size:   size,
							offset: 84 + ((6 + 12*uint32(count) + 7) &^ 7),
							reader: r,
						}, nil
					}
				}
			}
		} else if count, err := readUint16(r); err != nil {
			return nil, err
		} else if size, files, err := readIndex(r, count); err != nil {
			return nil, err
		} else {
			return &mixFile{
				files:  files,
				flags:  flags,
				size:   size,
				offset: 10 + 12*uint32(count),
				reader: r,
			}, nil
		}
	}
}

func (this *mixFile) OpenFile(i int) *io.SectionReader {
	info := this.files[i]
	return io.NewSectionReader(this.reader, int64(this.offset+info.offset), int64(info.size))
}

func (this *mixFile) ReadLmd(game gameId) error {
	lmdId := getLmdFileId(game)
	if fileIndex := this.files.indexById(lmdId); fileIndex == -1 {
		return nil
	} else if mapper, err := lmdRead(this.OpenFile(fileIndex)); err != nil {
		return err
	} else {
		for i := 0; i < len(this.files); i++ {
			if name, ok := mapper[this.files[i].id]; ok {
				this.files[i].name = name
			}
		}
		return nil
	}
}
