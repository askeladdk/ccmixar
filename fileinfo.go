package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type fileInfo interface {
	Name() string
	Size() int64
	Open() (io.ReadCloser, error)
}

type systemFileInfo struct {
	path string
	size int64
}

func (this *systemFileInfo) Name() string {
	return filepath.Base(this.path)
}

func (this *systemFileInfo) Size() int64 {
	return this.size
}

func (this *systemFileInfo) Open() (io.ReadCloser, error) {
	return os.Open(this.path)
}

type bufferFileInfo struct {
	name   string
	buffer bytes.Buffer
}

func (this *bufferFileInfo) Name() string {
	return this.name
}

func (this *bufferFileInfo) Size() int64 {
	return int64(this.buffer.Len())
}

func (this *bufferFileInfo) Open() (io.ReadCloser, error) {
	return ioutil.NopCloser(&this.buffer), nil
}

func readDirectory(dirname string) ([]fileInfo, error) {
	if fi1, err := ioutil.ReadDir(dirname); err != nil {
		return nil, err
	} else {
		var fi2 []fileInfo
		for _, fi := range fi1 {
			if !fi.IsDir() && strings.ToLower(fi.Name()) != lmdFilename {
				fi2 = append(fi2, &systemFileInfo{
					size: fi.Size(),
					path: path.Join(dirname, fi.Name()),
				})
			}
		}
		return fi2, nil
	}
}

func listFilesToPack(dirname string, database bool, gameId gameId) ([]fileInfo, error) {
	if files, err := readDirectory(dirname); err != nil {
		return nil, err
	} else if database {
		if lmb, err := lmdFile(gameId, files); err != nil {
			return nil, err
		} else {
			return append(files, lmb), nil
		}
	} else {
		return files, nil
	}
}
