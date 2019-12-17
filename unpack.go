package main

import (
	"io"

	"golang.org/x/crypto/blowfish"
)

type mixFileEntry struct {
	id     uint32
	offset uint32
	size   uint32
}

type mixFile struct {
	checksum []byte
	files    []mixFileEntry
	flags    uint32
	size     uint32
}

func readFileEntries(r io.Reader, count uint16) ([]mixFileEntry, error) {
	var entries []mixFileEntry
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

func readMixFile(r io.ReadSeeker) (*mixFile, error) {
	if count, err := readUint16(r); err != nil {
		return nil, err
	} else if count != 0 {
		if size, files, err := readIndex(r, count); err != nil {
			return nil, err
		} else {
			return &mixFile{
				flags: 0,
				size:  size,
				files: files,
			}, nil
		}
	} else if flags16, err := readUint16(r); err != nil {
		return nil, err
	} else {
		flags := uint32(flags16) << 16
		var size uint32
		var files []mixFileEntry
		var checksum []byte

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
					} else if size1, files1, err := readIndex(ecb, count); err != nil {
						return nil, err
					} else {
						size = size1
						files = files1
					}
				}
			}
		} else if size1, files1, err := readIndex(r, count); err != nil {
			return nil, err
		} else {
			size = size1
			files = files1
		}

		if (flags & flagChecksum) != 0 {
			checksum = make([]byte, 20)
			if _, err := r.Seek(-20, io.SeekEnd); err != nil {
				return nil, err
			} else if _, err := r.Read(checksum); err != nil {
				return nil, err
			}
		}

		return &mixFile{
			checksum: checksum,
			flags:    flags,
			files:    files,
			size:     size,
		}, nil
	}
}
