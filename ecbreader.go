package main

import (
	"crypto/cipher"
	"io"
)

type ecbReader struct {
	r io.Reader
	b cipher.Block
	t []byte
}

func newECBReader(r io.Reader, b cipher.Block) *ecbReader {
	return &ecbReader{
		r: r,
		b: b,
		t: nil,
	}
}

func (this *ecbReader) Read(p []byte) (int, error) {
	if len(this.t) < len(p) {
		blksz := this.b.BlockSize()
		n := (len(p) + blksz - 1) & ^(blksz - 1)
		t := make([]byte, n)
		if _, err := this.r.Read(t); err != nil {
			return 0, err
		} else {
			this.b.Decrypt(t, t)
			this.t = append(this.t, t...)
		}
	}

	copy(p, this.t[:len(p)])
	this.t = this.t[len(p):]
	return len(p), nil
}
