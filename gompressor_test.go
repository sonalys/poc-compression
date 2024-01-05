package gompressor

import (
	"bytes"
	"math"
	"os"
	"testing"
)

func Test_encoding(t *testing.T) {
	//	in := []byte{1, 1, 0, 1, 1}
	in, err := os.ReadFile("/bin/zsh")
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}

	block := compress(in, 7)
	compressed := encode(block)

	t.Run("serialization", func(t *testing.T) {
		decompressed, err := decode(compressed)
		if err != nil {
			t.Fatalf("failed to parse serialize: %s", err)
		}
		got := encode(decompressed)
		if !bytes.Equal(compressed, got) {
			t.Fatal("buffer is variating in serialization")
		}
	})

	t.Run("compression rate", func(t *testing.T) {
		ratio := float64(len(compressed)) / float64(len(in))
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
	var segmentCount int
	var highestRepeat int
	var highestGain int
	bestGroupSize := -1
	for groupSize := 2; groupSize < 30; groupSize++ {
		block := compress(in, uint16(groupSize))
		serialize := encode(block)
		if newSize := uint32(len(serialize)); newSize < bestSize {
			bestSize = newSize
			bestGroupSize = groupSize
			segmentCount = len(block.head)
			for _, entry := range block.head {
				if entry.repeat > uint16(highestRepeat) {
					highestRepeat = int(entry.repeat)
				}
				if gain := entry.getCompressionGains(); gain > highestGain {
					highestGain = gain
				}
			}
		}
	}
	ratio := float64(bestSize) / float64(len(in))
	t.Logf("best size %d group size %d. ratio %.2f", bestSize, bestGroupSize, ratio)
	t.Logf("segments count: %d highest repeat: %d highest compression gain: %d bytes", segmentCount, highestRepeat, highestGain)
	t.Fail()
}
