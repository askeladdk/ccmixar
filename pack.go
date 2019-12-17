package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"sort"

	"golang.org/x/crypto/blowfish"
)

const (
	flagChecksum  uint32 = 0x00010000
	flagEncrypted uint32 = 0x00020000
)

type filesAndIds struct {
	files []fileInfo
	ids   []uint32
}

func (this *filesAndIds) Len() int {
	return len(this.ids)
}

func (this *filesAndIds) Less(i, j int) bool {
	return int32(this.ids[i]) < int32(this.ids[j])
}

func (this *filesAndIds) Swap(i, j int) {
	this.files[i], this.files[j] = this.files[j], this.files[i]
	this.ids[i], this.ids[j] = this.ids[j], this.ids[i]
}

func writeIndex(w io.Writer, files []fileInfo, fileId fileId) error {
	count := uint16(len(files))
	size := uint32(0)

	for _, fi := range files {
		size += uint32(fi.Size())
	}

	if _, err := writeUint16(w, count); err != nil {
		return err
	} else if _, err := writeUint32(w, size); err != nil {
		return err
	}

	var ids []uint32
	for _, fi := range files {
		ids = append(ids, fileId(fi.Name()))
	}

	sort.Sort(&filesAndIds{
		files: files,
		ids:   ids,
	})

	for i := 1; i < len(ids); i++ {
		if ids[i-1] == ids[i] {
			return errors.New(fmt.Sprintf("ID collision %x on %s and %s", ids[i], files[i].Name(), files[i-1].Name()))
		}
	}

	offset := uint32(0)

	for i, fi := range files {
		id := ids[i]
		size := uint32(fi.Size())

		if _, err := writeUint32(w, id); err != nil {
			return err
		} else if _, err := writeUint32(w, offset); err != nil {
			return err
		} else if _, err := writeUint32(w, size); err != nil {
			return err
		}

		offset += size
	}

	return nil
}

func writeBody(w io.Writer, files []fileInfo) error {
	for _, fi := range files {
		if f, err := fi.Open(); err != nil {
			return err
		} else {
			defer f.Close()
			if _, err := io.Copy(w, f); err != nil {
				return err
			}
		}
	}

	return nil
}

func pack(w io.Writer, files []fileInfo, gameId gameId, flags uint32, keySource []byte) error {
	if gameId != gameId_CC1 {
		if _, err := writeUint32(w, flags); err != nil {
			return err
		}
	} else if flags != 0 {
		return errors.New("Game cc1 does not support flags.")
	}

	var fileId fileId
	if gameId <= gameId_RA1 {
		fileId = fileIdV1
	} else {
		fileId = fileIdV2
	}

	if (flags & flagEncrypted) != 0 {
		if _, err := w.Write(keySource); err != nil {
			return err
		}

		blowfishKey := blowfishKeyFromKeySource(keySource)
		if cipher, err := blowfish.NewCipher(blowfishKey); err != nil {
			return err
		} else {
			e := newEncryptWriter(w, cipher)
			if err := writeIndex(e, files, fileId); err != nil {
				return err
			} else if err := e.Flush(); err != nil {
				return err
			}
		}
	} else if err := writeIndex(w, files, fileId); err != nil {
		return err
	}

	if (flags & flagChecksum) != 0 {
		h := sha1.New()
		mw := io.MultiWriter(w, h)
		if err := writeBody(mw, files); err != nil {
			return err
		} else if _, err := w.Write(h.Sum(nil)); err != nil {
			return err
		}
	} else if err := writeBody(w, files); err != nil {
		return err
	}

	return nil
}
