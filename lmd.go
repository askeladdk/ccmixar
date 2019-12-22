package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

type gameId int

const (
	gameId_CC1 gameId = 0
	gameId_RA1 gameId = 1
	gameId_CC2 gameId = 2
	gameId_RA2 gameId = 5
	lmdHeader         = "XCC by Olaf van der Spek\x1a\x04\x17\x27\x10\x19\x80\x00"
)

const lmdFilename = "local mix database.dat"

func lmdWrite(gameId gameId, files []fileInfo) (fileInfo, error) {
	var b bytes.Buffer

	if _, err := fmt.Fprintf(&b, lmdHeader); err != nil {
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

func lmdRead(r io.ReadSeeker) (map[uint32]string, error) {
	var hdr [32]byte

	if _, err := r.Read(hdr[:]); err != nil {
		return nil, err
	} else if string(hdr[:]) != lmdHeader {
		return nil, errors.New("not a local mix database")
	} else if _, err := r.Seek(12, io.SeekCurrent); err != nil {
		return nil, err
	} else if gameid, err := readUint32(r); err != nil {
		return nil, err
	} else if _, err := r.Seek(4, io.SeekCurrent); err != nil {
		return nil, err
	} else {
		mapper := map[uint32]string{}
		fileId := getFileId(gameId(gameid))

		scanner := bufio.NewScanner(r)
		scanner.Split(scanZStrings)
		for scanner.Scan() {
			filename := scanner.Text()
			mapper[fileId(filename)] = filename
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		} else {
			return mapper, nil
		}
	}
}

func getLmdFileId(game gameId) uint32 {
	if game <= gameId_RA1 {
		return 0x54C2D545
	} else {
		return 0x366E051F
	}
}
