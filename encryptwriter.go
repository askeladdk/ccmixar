package main

import (
	"crypto/cipher"
	"io"
)

type encryptWriter struct {
	w io.Writer
	b cipher.Block
	t []byte
}

func newEncryptWriter(w io.Writer, b cipher.Block) *encryptWriter {
	return &encryptWriter{
		w: w,
		b: b,
		t: nil,
	}
}

func (this *encryptWriter) Write(p []byte) (int, error) {
	t := append(this.t, p...)
	blksz := this.b.BlockSize()
	remainder := len(t) % blksz
	nwritten := 0

	for i := 0; i < len(t)-remainder; i += blksz {
		blk := t[i : i+blksz]
		this.b.Encrypt(blk, blk)
		if n, err := this.w.Write(blk); err != nil {
			return nwritten + n, err
		} else {
			nwritten += n
		}
	}

	if remainder != 0 {
		this.t = t[len(t)-remainder:]
	} else {
		this.t = nil
	}

	return nwritten, nil
}

func (this *encryptWriter) Flush() error {
	if len(this.t) != 0 {
		blk := make([]byte, this.b.BlockSize())
		copy(blk, this.t)
		this.b.Encrypt(blk, blk)
		if _, err := this.w.Write(blk); err != nil {
			return err
		}
	}
	return nil
}
