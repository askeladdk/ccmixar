package main

import (
	"bytes"
	"crypto/cipher"
	"io"
)

type ecbReader struct {
	reader io.Reader
	block  cipher.Block
	buffer bytes.Buffer
}

func newECBReader(r io.Reader, b cipher.Block) *ecbReader {
	return &ecbReader{
		reader: r,
		block:  b,
	}
}

func (this *ecbReader) Read(p []byte) (int, error) {
	if this.buffer.Len() < len(p) {
		blksz := this.block.BlockSize()
		n := (len(p) + blksz - 1) & ^(blksz - 1)
		t := make([]byte, n)
		if _, err := this.reader.Read(t); err != nil {
			return 0, err
		} else {
			this.block.Decrypt(t, t)
			this.buffer.Write(t)
		}
	}

	return this.buffer.Read(p)
}
