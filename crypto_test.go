package main

import (
	"bytes"
	"testing"
)

func TestValidateKeys(t *testing.T) {
	privateKey.Precompute()
	if err := privateKey.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestBlowfishKeyFromSource(t *testing.T) {
	keySource := []byte{
		0xca, 0xd0, 0xb0, 0x1b, 0xfe, 0x3f, 0x3f, 0xb6,
		0xca, 0xc0, 0xbd, 0x8f, 0x40, 0xf0, 0xee, 0x85,
		0x6e, 0xe1, 0xda, 0x7a, 0xef, 0xb4, 0xd4, 0xbb,
		0x6a, 0xd8, 0x4b, 0x84, 0x26, 0x99, 0x6f, 0xfd,
		0x65, 0x97, 0xf2, 0x5f, 0xa4, 0x46, 0xdb, 0x47,
		0x88, 0x63, 0x4f, 0x2c, 0x14, 0x0b, 0x3c, 0xce,
		0xaa, 0xc4, 0x5c, 0xe4, 0x15, 0x86, 0x26, 0x5c,
		0x52, 0x3a, 0x80, 0xf8, 0xbe, 0x45, 0x40, 0x6a,
		0x66, 0xb4, 0xc5, 0xf6, 0xd0, 0x12, 0xe0, 0x43,
		0x44, 0x65, 0xc6, 0xe3, 0x9e, 0xf9, 0x43, 0x35,
	}

	expectedBlowfishKey := []byte{
		0x53, 0xb9, 0xb7, 0x6c, 0xec, 0x6c, 0x03, 0xb8,
		0x38, 0xb8, 0x6d, 0x11, 0x08, 0xac, 0x4a, 0x91,
		0x9d, 0x2f, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c,
		0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c, 0x0c,
		0x71, 0x6e, 0x94, 0xac, 0x2c, 0xac, 0xf0, 0x08,
		0x88, 0x08, 0xb5, 0x52, 0x4f, 0xec, 0x97, 0xd2,
		0x2a, 0x48, 0x05, 0x05, 0x05, 0x05, 0x05, 0x05,
	}

	bfkey := blowfishKeyFromKeySource(keySource)
	if !bytes.Equal(bfkey, expectedBlowfishKey) {
		t.Fatal("unexpected blowfish key")
	}
}
