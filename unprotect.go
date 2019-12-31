package main

import (
	"bytes"
	"io"

	"golang.org/x/crypto/blowfish"
)

func (this *mixFile) recoverLmdIndex() (uint32, uint32, bool) {
	var buf [0x400000]byte

	r := io.NewSectionReader(this.reader, int64(this.offset), int64(this.size))
	if _, err := r.Read(buf[len(buf)-32:]); err != nil {
		return 0, 0, false
	} else {
		for offset := uint32(0); ; offset += uint32(len(buf) - 32) {
			copy(buf[:32], buf[len(buf)-32:])
			if n, err := r.Read(buf[32:]); err != nil {
				return 0, 0, false
			} else {
				for i := uint32(0); i < uint32(n); i++ {
					if bytes.Compare(buf[i:i+32], []byte(lmdHeader)) == 0 {
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
	}
}

func (this *mixFile) RecoverLmd() {
	lmdid := getLmdFileId(this.game)
	for i, file := range this.files {
		if file.offset > this.size {
			if file.id == lmdid {
				if offset, size, found := this.recoverLmdIndex(); found {
					this.files[i].offset = offset
					this.files[i].size = size
				} else {
					this.files[i].offset = 0
					this.files[i].size = 0
				}
			} else {
				this.files[i].offset = 0
				this.files[i].size = 0
			}
		}
	}
}

func (this *mixFile) RewriteHeader(w io.Writer) error {
	writeIndex := func(w io.Writer) error {
		writeUint16(w, uint16(len(this.files)))
		writeUint32(w, this.size)

		for _, file := range this.files {
			writeUint32(w, file.id)
			writeUint32(w, file.offset)
			writeUint32(w, file.size)
		}

		return nil
	}

	if this.game != gameId_CC1 {
		writeUint32(w, this.flags)
	}

	if this.game != gameId_CC1 && (this.flags&flagEncrypted) != 0 {
		w.Write(this.keysrc)
		bfkey := blowfishKeyFromKeySource(this.keysrc)
		if b, err := blowfish.NewCipher(bfkey); err != nil {
			return err
		} else {
			w2 := newEcbWriter(w, b)
			if err := writeIndex(w2); err != nil {
				return err
			} else {
				return w2.Flush()
			}
		}
	} else {
		return writeIndex(w)
	}
}
