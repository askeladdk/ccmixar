package main

import (
	"bytes"
	"encoding/binary"
	"io"

	"golang.org/x/crypto/blowfish"
)

func (mix *mixFile) recoverLmdIndex() (uint32, uint32, bool) {
	var buf [0x400000]byte

	r := io.NewSectionReader(mix.reader, int64(mix.offset), int64(mix.size))
	if _, err := r.Read(buf[len(buf)-32:]); err != nil {
		return 0, 0, false
	}
	for offset := uint32(0); ; offset += uint32(len(buf) - 32) {
		copy(buf[:32], buf[len(buf)-32:])
		n, err := r.Read(buf[32:])
		if err != nil {
			return 0, 0, false
		}
		for i := uint32(0); i < uint32(n); i++ {
			if bytes.Equal(buf[i:i+32], []byte(lmdHeader)) {
				lmdofs := offset + i
				if _, err := r.Seek(int64(lmdofs+32), io.SeekStart); err != nil {
					return 0, 0, false
				} else if lmdsize, err := readUint32(r); err != nil {
					return 0, 0, false
				} else {
					return lmdofs, lmdsize, true
				}
			}
		}
	}
}

func (mix *mixFile) RecoverLmd() {
	lmdID := getLmdFileID(mix.game)
	for i, file := range mix.files {
		if file.offset > mix.size {
			if file.id == lmdID {
				if offset, size, found := mix.recoverLmdIndex(); found {
					mix.files[i].offset = offset
					mix.files[i].size = size
				} else {
					mix.files[i].offset = 0
					mix.files[i].size = 0
				}
			} else {
				mix.files[i].offset = 0
				mix.files[i].size = 0
			}
		}
	}
}

func (mix *mixFile) RewriteHeader(w io.Writer) error {
	writeIndex := func(w io.Writer) error {
		buf := make([]byte, 6+12*len(mix.files))
		binary.LittleEndian.PutUint16(buf[0:], uint16(len(mix.files)))
		binary.LittleEndian.PutUint32(buf[4:], mix.size)
		for i, file := range mix.files {
			binary.LittleEndian.PutUint32(buf[6+12*i+0:], file.id)
			binary.LittleEndian.PutUint32(buf[6+12*i+4:], file.offset)
			binary.LittleEndian.PutUint32(buf[6+12*i+8:], file.size)
		}

		_, err := w.Write(buf)
		return err
	}

	if mix.game != gameCC1 {
		if _, err := writeUint32(w, mix.flags); err != nil {
			return err
		}
	}

	if mix.game != gameCC1 && (mix.flags&flagEncrypted) != 0 {
		if _, err := w.Write(mix.keysrc); err != nil {
			return err
		}
		bfkey := blowfishKeyFromKeySource(mix.keysrc)
		b, err := blowfish.NewCipher(bfkey)
		if err != nil {
			return err
		}
		w2 := newEcbWriter(w, b)
		if err := writeIndex(w2); err != nil {
			return err
		}
		return w2.Flush()
	}
	return writeIndex(w)
}
