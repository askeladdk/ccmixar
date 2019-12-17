package main

import (
	"bytes"
	"fmt"
)

type gameId int

const (
	gameId_CC1 gameId = 0
	gameId_RA1 gameId = 1
	gameId_CC2 gameId = 2
	gameId_RA2 gameId = 5
)

const lmdFilename = "local mix database.dat"

func lmdFile(gameId gameId, files []fileInfo) (fileInfo, error) {
	var b bytes.Buffer

	if _, err := fmt.Fprintf(&b, "XCC by Olaf van der Spek\x1a\x04\x17\x27\x10\x19\x80\x00"); err != nil {
		return nil, err
	}

	size := uint32(52 + 1 + len(lmdFilename))
	var filesToKeep []fileInfo

	for _, f := range files {
		if _, ok := filenameIsId(f.Name()); !ok {
			filesToKeep = append(filesToKeep, f)
			size += uint32(1 + len(f.Name()))
		}
	}

	for _, v := range []uint32{size, 0, 0, uint32(gameId), 1 + uint32(len(files))} {
		if _, err := writeUint32(&b, v); err != nil {
			return nil, err
		}
	}

	for _, f := range filesToKeep {
		if _, err := fmt.Fprintf(&b, "%s\x00", f.Name()); err != nil {
			return nil, err
		}
	}

	if _, err := fmt.Fprintf(&b, "%s\x00", lmdFilename); err != nil {
		return nil, err
	}

	return &bufferFileInfo{
		name:   lmdFilename,
		buffer: b,
	}, nil
}
