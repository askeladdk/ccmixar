package main

import (
	"encoding/binary"
	"io"
)

func writeUint16(w io.Writer, v uint16) (int, error) {
	b := [2]byte{}
	binary.LittleEndian.PutUint16(b[:], v)
	return w.Write(b[:])
}

func writeUint32(w io.Writer, v uint32) (int, error) {
	b := [4]byte{}
	binary.LittleEndian.PutUint32(b[:], v)
	return w.Write(b[:])
}

func byteswap(b []byte) {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
}
