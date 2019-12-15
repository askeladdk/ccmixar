package main

import (
	"hash"
	"io"
)

type hashWriter struct {
	w io.Writer
	h hash.Hash
}

func newHashWriter(w io.Writer, h hash.Hash) *hashWriter {
	return &hashWriter{w: w, h: h}
}

func (this *hashWriter) Write(p []byte) (int, error) {
	this.h.Write(p)
	return this.w.Write(p)
}

func (this *hashWriter) Sum(b []byte) []byte {
	return this.h.Sum(b)
}
