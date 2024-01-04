package gompressor

import (
	"bytes"
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
		block := compress(in, 1)

		rec := decompress(block)

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
		block := compress(in, 6)
		serialized := block.serialize()

		if len(serialized) > len(in) {
			t.Fatalf("compression is bigger: original %d bytes compressed %d bytes. ratio %.2f", len(in), len(serialized), float64(len(serialized))/float64(len(in)))
		}
		// compressedSize := len(entries)*int(entrySize) + len(out)
		// if len(in) < compressedSize {
		// 	t.Errorf("compressed size %d should be smaller than original %d", len(in), compressedSize)
		// }
		// t.Logf("original: %d compressed: %d ratio: %.2f", len(in), compressedSize, float64(compressedSize)/float64(len(in)))
		rec := decompress(block)
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

func Test_serializeParse(t *testing.T) {
	t.Run("should return to same state", func(t *testing.T) {
		in := []byte{1, 1, 0, 1, 1, 2, 2, 0, 3, 3, 0}
		block := compress(in, 2)
		serialize := block.serialize()
		ratio := float64(len(serialize)) / float64(len(in))
		if ratio > 1 {
			t.Errorf("compression increased file size. ratio: %.2f", ratio)
		}
		out := decompress(block)
		for i := range out {
			if out[i] != in[i] {
				t.Fatalf("invalid reconstruction at pos %d expected %d got %d", i, in[i], out[i])
			}
		}
		resp, err := parse(serialize)
		if err != nil {
			t.Fatalf("failed to parse serialize: %s", err)
		}
		serialize2 := resp.serialize()

		if !bytes.Equal(serialize, serialize2) {
			t.Fatal("buffer is variating in serialization")
		}
	})
}
