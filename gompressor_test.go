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
		pos, got := calculateSmallestDelta(in, 0, windowSize)
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

	expBlock := Compress(in)
	expSerialized := Encode(expBlock)
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

func Test_bestMinSize(t *testing.T) {
	// in, err := os.ReadFile("/home/raicon/Pictures/Screenshot_20240105_145006.png")
	in, err := os.ReadFile("/bin/zsh")
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	// stats := CountRepetitions(in)
	// t.Logf("%v", stats)
	var segmentCount int
	var highestRepeat int
	var highestGain, highestLoss int64 = 0, math.MaxInt64
	block := Compress(in)
	serialize := Encode(block)
	newSize := int64(len(serialize))
	// segmentCount = len(block.Segments)

	out := Decompress(block)
	require.Equal(t, len(in), len(out), "input and output are different")
	for i := range in {
		if in[i] != out[i] {
			window := 20
			require.Equal(t, in[i], out[i], "in[%d] != out[%d]\n%v\n%v", i, i, in[i:i+window], out[i:i+window])
		}
	}

	block.ForEach(func(entry *Segment) {
		if entry.Repeat > uint16(highestRepeat) {
			highestRepeat = int(entry.Repeat)
		}
		if gain := entry.GetCompressionGains(); gain > highestGain {
			highestGain = gain
		} else if gain < highestLoss {
			highestLoss = gain
		}
	})

	ratio := float64(newSize) / float64(len(in))
	t.Logf(`
byte ratio				%.2f (%d / %d)
compressed: 			%d bytes
segments count: 	%d
highest repeat: 	%d
highest delta: 		%d bytes
smallest delta: 		%d bytes
`,
		ratio, newSize, int64(len(in)), int64(len(in))-newSize, segmentCount, highestRepeat, highestGain, highestLoss)
}
