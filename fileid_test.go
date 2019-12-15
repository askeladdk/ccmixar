package main

import "testing"

func Test_fileIdV1(t *testing.T) {
	tests := []struct {
		name string
		id   uint32
	}{
		{"local mix database.dat", 0x54C2D545},
		{"rules.ini", 0xB1C3B238},
		{"harv.shp", 0xFCECD5BE},
		{"CAFEBABE", 0xCAFEBABE},
		{"image.pcx", 0xA3A59207},
		{"5tnk.shp", 0xE6E4FB98},
		{"scenario.ini", 0x20F5FAFD},
	}

	for _, test := range tests {
		if fileIdV1(test.name) != test.id {
			t.Fatal(test.name)
		}
	}
}

func Test_fileIdV2(t *testing.T) {
	tests := []struct {
		name string
		id   uint32
	}{
		{"local mix database.dat", 0x366E051F},
		{"rules.ini", 0xF025A96C},
		{"harv.vxl", 0xAEE7BB83},
		{"CAFEBABE", 0xCAFEBABE},
	}

	for _, test := range tests {
		if fileIdV2(test.name) != test.id {
			t.Fatal(test.name)
		}
	}
}
