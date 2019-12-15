package main

import (
	"io"
	"os"
	"testing"
)

func Test_lmdFile(t *testing.T) {
	if files, err := readDirectory("./test/files"); err != nil {
		t.Fatal(err)
	} else if lmd, err := lmdFile(gameId_CC1, files); err != nil {
		t.Fatal(err)
	} else if f, err := os.OpenFile("./test/local mix database.dat", os.O_CREATE|os.O_RDWR, 0644); err != nil {
		t.Fatal(err)
	} else if r, err := lmd.Open(); err != nil {
		t.Fatal(err)
	} else if _, err := io.Copy(f, r); err != nil {
		t.Fatal(err)
	} else if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
