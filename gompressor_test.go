package gompressor

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math"
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
		in := []byte{0, 1, 1, 2, 2, 1, 1, 2, 2, 3}
		block := compress(in, 2)

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
	in, err := os.ReadFile("/bin/zsh")
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}

	block := compress(in, 7)
	serialize := block.serialize()

	t.Run("serialization", func(t *testing.T) {
		resp, err := parse(serialize)
		if err != nil {
			t.Fatalf("failed to parse serialize: %s", err)
		}
		got := resp.serialize()
		if !bytes.Equal(serialize, got) {
			t.Fatalf("buffer is variating in serialization.\nexp:\n%v\ngot\n%v", serialize, got)
		}
	})

	t.Run("compression rate", func(t *testing.T) {
		ratio := float64(len(serialize)) / float64(len(in))
		if ratio > 1 {
			t.Errorf("compression increased file size. ratio: %.2f", ratio)
		}
	})

	t.Run("reconstruction", func(t *testing.T) {
		out := decompress(block)
		for i := range out {
			if out[i] != in[i] {
				t.Logf("exp:\n%v\ngot:\n%v", in[i-10:i+10], out[i-10:i+10])
				t.Fatalf("invalid reconstruction at pos %d expected %d got %d", i, in[i], out[i])
			}
		}
	})
}

func Test_bestMinSize(t *testing.T) {
	in, err := os.ReadFile("/bin/zsh")
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	var bestSize uint32 = math.MaxUint32
	bestGroupSize := -1
	for groupSize := 2; groupSize < 20; groupSize++ {
		block := compress(in, uint16(groupSize))
		serialize := block.serialize()
		if newSize := uint32(len(serialize)); newSize < bestSize {
			bestSize = newSize
			bestGroupSize = groupSize
		}
	}
	ratio := float64(bestSize) / float64(len(in))
	t.Logf("best size %d group size %d. ratio %.2f", bestSize, bestGroupSize, ratio)
	t.Fail()
}
