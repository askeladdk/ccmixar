package main

import (
	"bytes"
	"crypto/sha1"
	"testing"
)

func Test_hashWriter(t *testing.T) {
	var buf bytes.Buffer
	correct := sha1.Sum([]byte("hello world"))

	checksum := newHashWriter(&buf, sha1.New())

	if _, err := checksum.Write([]byte("hello world")); err != nil {
		t.Fatal(err)
	} else if bytes.Compare(correct[:], checksum.Sum(nil)) != 0 {
		t.Fatalf("incorrect hash")
	} else if buf.String() != "hello world" {
		t.Fatalf("incorrect string")
	}
}
