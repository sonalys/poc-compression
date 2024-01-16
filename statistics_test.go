package gompressor

import (
	"math"
	"os"
	"testing"
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
