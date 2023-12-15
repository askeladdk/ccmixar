package main

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"io/fs"
	"os"
)

//go:embed cc1gmd.csv cc2gmd.csv ra1gmd.csv ra2gmd.csv
var gmdfs embed.FS

func gmdRead(filename string, gameid gameID) (map[uint32]string, error) {
	var f fs.File

	if filename == "" {
		emf, err := gmdfs.Open(fmt.Sprintf("%sgmd.csv", gameid))
		if err != nil {
			return nil, err
		}
		f = emf
	} else {
		osf, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		f = osf
	}

	defer f.Close()

	fileid := getFileID(gameid)

	mapper := make(map[uint32]string)

	c := csv.NewReader(f)
	c.ReuseRecord = true
	c.Comma = '\t'
	c.FieldsPerRecord = -1

	for {
		rec, err := c.Read()
		if err == io.EOF {
			return mapper, nil
		} else if err != nil {
			return nil, err
		}
		name := rec[0]
		mapper[fileid(name)] = name
	}
}
