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

func (fids *filesAndIds) Len() int {
	return len(fids.ids)
}

func (fids *filesAndIds) Less(i, j int) bool {
	return int32(fids.ids[i]) < int32(fids.ids[j])
}

func (fids *filesAndIds) Swap(i, j int) {
	fids.files[i], fids.files[j] = fids.files[j], fids.files[i]
	fids.ids[i], fids.ids[j] = fids.ids[j], fids.ids[i]
}

func writeIndex(w io.Writer, files []fileInfo, fileID fileID) error {
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
		ids = append(ids, fileID(fi.Name()))
	}

	sort.Sort(&filesAndIds{
		files: files,
		ids:   ids,
	})

	for i := 1; i < len(ids); i++ {
		if ids[i-1] == ids[i] {
			return fmt.Errorf("ID collision %x on %s and %s", ids[i], files[i].Name(), files[i-1].Name())
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
		f, err := fi.Open()
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(w, f); err != nil {
			return err
		}
	}
	return nil
}

func pack(w io.Writer, files []fileInfo, game gameID, flags uint32, keySource []byte) error {
	if game != gameCC1 {
		if _, err := writeUint32(w, flags); err != nil {
			return err
		}
	} else if flags != 0 {
		return errors.New("game cc1 does not support flags")
	}

	fileID := getFileID(game)

	if (flags & flagEncrypted) != 0 {
		if _, err := w.Write(keySource); err != nil {
			return err
		}

		blowfishKey := blowfishKeyFromKeySource(keySource)
		cipher, err := blowfish.NewCipher(blowfishKey)
		if err != nil {
			return err
		}
		e := newEcbWriter(w, cipher)
		if err := writeIndex(e, files, fileID); err != nil {
			return err
		} else if err := e.Flush(); err != nil {
			return err
		}
	} else if err := writeIndex(w, files, fileID); err != nil {
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
