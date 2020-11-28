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

func (w *ecbWriter) Write(p []byte) (int, error) {
	blksz := w.block.BlockSize()

	w.buffer.Write(p)

	for w.buffer.Len() >= blksz {
		blk := make([]byte, blksz)
		w.buffer.Read(blk)
		w.block.Encrypt(blk, blk)
		if _, err := w.writer.Write(blk); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

func (w *ecbWriter) Flush() error {
	if w.buffer.Len() != 0 {
		blk := make([]byte, w.block.BlockSize())
		w.buffer.Read(blk)
		w.block.Encrypt(blk, blk)
		if _, err := w.writer.Write(blk); err != nil {
			return err
		}
	}
	return nil
}
