package gompressor

import (
	"bytes"
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

	resp := countBytes(in)
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
	in := []byte{0, 1, 1, 2, 2, 0, 1, 1, 2, 2, 0}
	// in, err := os.ReadFile("/bin/zsh")
	// if err != nil {
	// 	t.Fatalf("failed to read file: %s", err)
	// }

	block := Compress(in, 2)
	expectedSerialization := encode(block)
	decoded, err := decode(expectedSerialization)
	require.NoError(t, err)

	t.Run("segment composition", func(t *testing.T) {
		exp := [][]byte{
			{0},
			{1},
			{2},
		}

		got := [][]byte{}
		for i := range block.head {
			got = append(got, block.head[i].buffer)
		}
		require.Equal(t, exp, got)
	})

	t.Run("decoding", func(t *testing.T) {
		require.Equal(t, len(block.head), len(decoded.head))
		for i := range block.head {
			cur := block.head[i]
			cur.next = nil
			cur.previous = nil
			cur.pos = nil
			require.Equal(t, block.head[i], decoded.head[i], "segment %d is different", i)
		}
	})

	t.Run("encoding", func(t *testing.T) {
		if err != nil {
			t.Fatalf("failed to parse serialize: %s", err)
		}
		got := encode(decoded)
		if !bytes.Equal(expectedSerialization, got) {
			t.Fatal("buffer is variating in serialization")
		}
	})

	t.Run("compression rate", func(t *testing.T) {
		ratio := float64(len(expectedSerialization)) / float64(len(in))
		if ratio > 1 {
			t.Errorf("compression increased file size. ratio: %.2f", ratio)
		}
	})

	t.Run("reconstruction", func(t *testing.T) {
		out := Decompress(block)
		require.Equal(t, in, out)
	})
}

func Test_bestMinSize(t *testing.T) {
	in, err := os.ReadFile("/bin/zsh")
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	var bestSize int64 = math.MaxInt64
	var segmentCount int
	var highestRepeat int
	var highestGain int64
	bestGroupSize := -1
	for groupSize := 2; groupSize < 30; groupSize++ {
		block := Compress(in, uint16(groupSize))
		serialize := encode(block)
		newSize := int64(len(serialize))
		if newSize >= bestSize {
			continue
		}
		bestSize = newSize
		bestGroupSize = groupSize
		segmentCount = len(block.head)
		for _, entry := range block.head {
			if entry.repeat > uint16(highestRepeat) {
				highestRepeat = int(entry.repeat)
			}
			if gain := entry.GetCompressionGains(); gain > highestGain {
				highestGain = gain
			}
		}
	}
	ratio := float64(bestSize) / float64(len(in))
	t.Logf(`
byte ratio			%.2f (%d / %d)
compressed: 		%d bytes
best minSize		%d
segments count: %d
highest repeat: %d
highest gain: 	%d bytes`, ratio, bestSize, int64(len(in)), int64(len(in))-bestSize, bestGroupSize, segmentCount, highestRepeat, highestGain)
	t.Fail()
}
