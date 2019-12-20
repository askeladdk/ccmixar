package main

import (
	"bytes"
	"crypto/cipher"
	"io"
)

type ecbWriter struct {
	writer io.Writer
	block  cipher.Block
	buffer bytes.Buffer
}

func newEcbWriter(w io.Writer, b cipher.Block) *ecbWriter {
	return &ecbWriter{
		writer: w,
		block:  b,
	}
}

func (this *ecbWriter) Write(p []byte) (int, error) {
	blksz := this.block.BlockSize()

	this.buffer.Write(p)

	for this.buffer.Len() >= blksz {
		blk := make([]byte, blksz)
		this.buffer.Read(blk)
		this.block.Encrypt(blk, blk)
		if _, err := this.writer.Write(blk); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

func (this *ecbWriter) Flush() error {
	if this.buffer.Len() != 0 {
		blk := make([]byte, this.block.BlockSize())
		this.buffer.Read(blk)
		this.block.Encrypt(blk, blk)
		if _, err := this.writer.Write(blk); err != nil {
			return err
		}
	}
	return nil
}
