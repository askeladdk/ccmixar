package main

import "testing"

func TestCommandPack(t *testing.T) {
	if err := commandPack([]string{"-dir", "./test/files", "-mix", "./test/mytest.mix", "-game", "ra2", "-database", "-checksum"}); err != nil {
		t.Fatal(err)
	}
}
