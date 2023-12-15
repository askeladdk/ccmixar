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

func (info *systemFileInfo) Name() string {
	return filepath.Base(info.path)
}

func (info *systemFileInfo) Size() int64 {
	return info.size
}

func (info *systemFileInfo) Open() (io.ReadCloser, error) {
	return os.Open(info.path)
}

type bufferFileInfo struct {
	name   string
	buffer bytes.Buffer
}

func (info *bufferFileInfo) Name() string {
	return info.name
}

func (info *bufferFileInfo) Size() int64 {
	return int64(info.buffer.Len())
}

func (info *bufferFileInfo) Open() (io.ReadCloser, error) {
	return io.NopCloser(&info.buffer), nil
}

func readDirectory(dirname string) ([]fileInfo, error) {
	fi1, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
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

func listFilesToPack(dirname string, database bool, gameID gameID) ([]fileInfo, error) {
	if files, err := readDirectory(dirname); err != nil {
		return nil, err
	} else if database {
		lmb, err := lmdWrite(gameID, files)
		if err != nil {
			return nil, err
		}
		return append(files, lmb), nil
	} else {
		return files, nil
	}
}
