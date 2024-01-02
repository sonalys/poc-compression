package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"slices"
	"testing"
)

func Test_countBytes(t *testing.T) {
	input := []byte{0, 1, 2, 3, 2, 1, 0}

	count := countBytes(input)

	exp := [256]uint{
		2, 2, 2, 1,
	}

	for i := range exp {
		if exp[i] != count[i] {
			t.Fail()
		}
	}
}

func Test_RandomFile(t *testing.T) {
	buffer := make([]byte, 4000)
	rand.Read(buffer)
	count := countBytes(buffer)
	slices.Sort(count)
	fmt.Print(count)
}

func Test_compressRepetition(t *testing.T) {
	t.Run("small sequences", func(t *testing.T) {
		in := []byte{0, 1, 1, 1, 0, 2, 2, 0, 3, 3, 0}
		entries, out := compressRepetition(in, 1)
		if len(entries) != 3 {
			t.Error("invalid entries")
		}
		if len(out) != 4 {
			t.Error("invalid out")
		}

		rec := reconstruct(out, entries)

		if len(rec) != len(in) {
			t.Fatal("failed reconstruct size")
		}

		for i := range rec {
			if rec[i] != in[i] {
				t.Fatal("failed reconstruct value")
			}
		}
	})

	t.Run("compress bin/zsh", func(t *testing.T) {
		in, err := os.ReadFile("/bin/zsh")
		if err != nil {
			t.Fatalf("failed to read file: %s", err)
		}
		entries, out := compressRepetition(in, entrySize)
		compressedSize := len(entries)*int(entrySize) + len(out)
		if len(in) < compressedSize {
			t.Errorf("compressed size %d should be smaller than original %d", len(in), compressedSize)
		}
		t.Logf("original: %d compressed: %d ratio: %.2f", len(in), compressedSize, float64(compressedSize)/float64(len(in)))
		rec := reconstruct(out, entries)
		if len(rec) != len(in) {
			t.Error("invalid reconstruction size")
		}
		for i := range rec {
			if rec[i] != in[i] {
				t.Fatal("invalid reconstruction")
			}
		}
	})
}
