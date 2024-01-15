package gompressor

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_countBytes(t *testing.T) {
	in, err := os.ReadFile("/home/raicon/Pictures/Screenshot_20240105_145006.png")
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}

	resp := CountBytes(in)
	t.Logf("%v", resp)
	t.Fail()
}

func Test_averageByteShift(t *testing.T) {
	in, err := os.ReadFile("/home/raicon/Pictures/Screenshot_20240105_145006.png")
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}

	var smallestByteShift float64 = math.MaxFloat64

	var bestWindow, bestPos uint32

	for windowSize := uint32(100); windowSize > 9; windowSize-- {
		pos, got := CalculateSmallestDelta(in, 0, windowSize)
		if got < smallestByteShift {
			smallestByteShift = got
			bestWindow = windowSize
			bestPos = pos
		}
	}

	t.Logf("byte shift: %v window %d", smallestByteShift, bestWindow)
	t.Logf("input[%d] = %v", bestPos, in[bestPos:bestPos+bestWindow])
	t.Fail()
}

func Test_encoding(t *testing.T) {
	in := []byte{255, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 255}
	// in, err := os.ReadFile("/bin/zsh")
	// if err != nil {
	// 	t.Fatalf("failed to read file: %s", err)
	// }

	expBlock := NewBlock(in)
	expSerialized := Compress(expBlock)
	gotBlock, err := Decode(expSerialized)
	require.NoError(t, err)

	t.Run("reconstruction", func(t *testing.T) {
		got := Decompress(expBlock)
		require.Equal(t, in, got)
	})

	t.Run("deserialized reconstruction", func(t *testing.T) {
		out := Decompress(gotBlock)
		require.Equal(t, in, out)
	})

	t.Run("compression rate", func(t *testing.T) {
		ratio := float64(len(expSerialized)) / float64(len(in))
		if ratio > 1 {
			t.Errorf("compression increased file size. ratio: %.2f", ratio)
		}
	})
}
