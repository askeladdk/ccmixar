package main

import (
	"crypto/cipher"
	"io"
)

type ecbWriter struct {
	w io.Writer
	b cipher.Block
	t []byte
}

func newEcbWriter(w io.Writer, b cipher.Block) *ecbWriter {
	return &ecbWriter{
		w: w,
		b: b,
		t: nil,
	}
}

func (this *ecbWriter) Write(p []byte) (int, error) {
	t := append(this.t, p...)
	blksz := this.b.BlockSize()
	remainder := len(t) % blksz

	for i := 0; i < len(t)-remainder; i += blksz {
		blk := t[i : i+blksz]
		this.b.Encrypt(blk, blk)
		if _, err := this.w.Write(blk); err != nil {
			return 0, err
		}
	}

	if remainder != 0 {
		this.t = t[len(t)-remainder:]
	} else {
		this.t = nil
	}

	return len(p), nil
}

func (this *ecbWriter) Flush() error {
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
