package segments

import (
	"math"

	"github.com/sonalys/gompressor/compression"
	ll "github.com/sonalys/gompressor/linkedlist"
)

func recalculateWindowMaskGain(in []byte, mask byte, pos, size int) (byte, int, bool) {
	mask, _, _, sizeMask := compression.MaskRegisterByte(mask, in[pos])
	if sizeMask == 8 {
		return mask, math.MinInt, false
	}
	compressedSize := (size*sizeMask + 8 - 1) / 8
	gain := size - compressedSize
	return mask, gain, true
}

func getBestEnd(in []byte, start, bestEnd, bestGain int) (int, int) {
	inLen := len(in)
	var mask byte
	var gain int
	var maskSize int
	var ok bool
	for i := start; i <= bestEnd; i++ {
		mask, _, _, maskSize = compression.MaskRegisterByte(mask, in[i])
		if maskSize == 8 {
			return bestEnd, gain
		}
	}
	for newEnd := bestEnd + 1; bestEnd < inLen-1; newEnd++ {
		mask, gain, ok = recalculateWindowMaskGain(in, mask, newEnd, newEnd-start)
		if !ok || gain < bestGain {
			break
		}
		bestEnd = newEnd
		bestGain = gain
	}
	return bestEnd, bestGain
}

func getBestStart(in []byte, bestStart, end, bestGain int, minSize int) (int, int) {
	for newStart := bestStart + 1; bestStart+minSize < end; newStart++ {
		var mask byte
		var maskSize int
		var gain int
		var ok bool
		// We are removing the first byte, so we always have to recalculate the masks.
		for i := newStart; i <= end; i++ {
			mask, _, _, maskSize = compression.MaskRegisterByte(mask, in[i])
			if maskSize == 8 {
				return bestStart, bestGain
			}
		}
		_, gain, ok = recalculateWindowMaskGain(in, mask, newStart, end-newStart)
		if !ok {
			return bestStart, bestGain
		}
		if gain < bestGain {
			break
		}
		bestStart = newStart
		bestGain = gain
	}
	return bestStart, bestGain
}

func CreateMaskedSegments(in []byte) (*ll.LinkedList[Segment], []byte) {
	list := &ll.LinkedList[Segment]{}
	inLen := len(in)
	const minSize = 6
	for i := 0; i < inLen-minSize-1; i++ {
		bestGain := math.MinInt
		bestStart := i
		bestEnd := i
		if bestEnd > inLen {
			bestEnd = inLen
		}
		for {
			prevGain := bestGain
			if newEnd, gain := getBestEnd(in, bestStart, bestEnd, bestGain); gain > bestGain {
				bestEnd = newEnd
				bestGain = gain
			}
			if newStart, gain := getBestStart(in, bestStart, bestEnd, bestGain, minSize); gain > bestGain {
				bestStart = newStart
				bestGain = gain
			}
			if prevGain == bestGain {
				break
			}
		}
		// println(bestStart, bestEnd, bestGain)
		if seg := NewMaskedSegment(in[bestStart:bestEnd+1], bestStart); seg.GetCompressionGains() > 0 {
			list.AppendValue(seg)
			i = bestEnd
		}
	}
	return list, FillSegmentGaps(in, list)
}
