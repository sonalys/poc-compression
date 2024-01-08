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

	block := Compress(in)
	expectedSerialization := Encode(block)
	decoded, err := Decode(expectedSerialization)
	require.NoError(t, err)

	t.Run("segment composition", func(t *testing.T) {
		exp := [][]byte{
			{0},
			{1},
			{2},
		}

		got := [][]byte{}
		for i := range block.Segments {
			got = append(got, block.Segments[i].Buffer)
		}
		require.Equal(t, exp, got)
	})

	t.Run("decoding", func(t *testing.T) {
		require.Equal(t, len(block.Segments), len(decoded.Segments))
		for i := range block.Segments {
			cur := block.Segments[i]
			cur.Next = nil
			cur.Previous = nil
			cur.Pos = nil
			require.Equal(t, block.Segments[i], decoded.Segments[i], "segment %d is different", i)
		}
	})

	t.Run("encoding", func(t *testing.T) {
		if err != nil {
			t.Fatalf("failed to parse serialize: %s", err)
		}
		got := Encode(decoded)
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
	in, err := os.ReadFile("/home/raicon/Pictures/Screenshot_20240105_145006.png")
	//in, err := os.ReadFile("/bin/zsh")
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	var segmentCount int
	var highestRepeat int
	var highestGain int64
	block := Compress(in)
	serialize := Encode(block)
	newSize := int64(len(serialize))
	segmentCount = len(block.Segments)

	for _, entry := range block.Segments {
		if entry.Repeat > uint16(highestRepeat) {
			highestRepeat = int(entry.Repeat)
		}
		if gain := entry.GetCompressionGains(); gain > highestGain {
			highestGain = gain
		}
	}

	ratio := float64(newSize) / float64(len(in))
	t.Logf(`
byte ratio			%.2f (%d / %d)
compressed: 		%d bytes
segments count: %d
highest repeat: %d
highest gain: 	%d bytes`, ratio, newSize, int64(len(in)), int64(len(in))-newSize, segmentCount, highestRepeat, highestGain)
	t.Fail()
}
