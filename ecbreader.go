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

func (r *ecbReader) Read(p []byte) (int, error) {
	if r.buffer.Len() < len(p) {
		blksz := r.block.BlockSize()
		n := (len(p) + blksz - 1) & ^(blksz - 1)
		t := make([]byte, n)
		if _, err := r.reader.Read(t); err != nil {
			return 0, err
		}
		r.block.Decrypt(t, t)
		r.buffer.Write(t)
	}

	return r.buffer.Read(p)
}
