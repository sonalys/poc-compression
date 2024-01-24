package segments

import (
	"math"

	"github.com/sonalys/gompressor/compression"
	ll "github.com/sonalys/gompressor/linkedlist"
)

func recalculateWindowMaskGain(in []byte, mask, notMask byte, pos, size int) (byte, byte, int, bool) {
	mask |= in[pos]
	notMask |= ^in[pos]
	if mask == 0xff && notMask == 0xff {
		return mask, notMask, math.MinInt, false
	}
	sizeMask, sizeNotMask := compression.Count1Bits(mask), compression.Count1Bits(notMask)
	if sizeNotMask < sizeMask {
		sizeMask = sizeNotMask
	}
	compressedSize := (size*sizeMask + 8 - 1) / 8
	gain := size - compressedSize
	return mask, notMask, gain, true
}

func getBestEnd(in []byte, start, bestEnd, bestGain int) (int, int) {
	inLen := len(in)
	var mask byte
	var notMask byte
	var gain int
	var ok bool
	for i := start; i <= bestEnd; i++ {
		mask |= in[i]
		notMask |= ^in[i]
		if mask == 0xff && notMask == 0xff {
			return bestEnd, bestGain
		}
	}
	for newEnd := bestEnd + 1; bestEnd < inLen-1; newEnd++ {
		mask, notMask, gain, ok = recalculateWindowMaskGain(in, mask, notMask, newEnd, newEnd-start)
		if !ok {
			return bestEnd, bestGain
		}
		if gain <= bestGain {
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
		var notMask byte
		var gain int
		var ok bool
		// We are removing the first byte, so we always have to recalculate the masks.
		for i := newStart; i <= end; i++ {
			mask |= in[i]
			notMask |= ^in[i]
			if mask == 0xff && notMask == 0xff {
				return bestStart, bestGain
			}
		}
		_, _, gain, ok = recalculateWindowMaskGain(in, mask, notMask, newStart, end-newStart)
		if !ok {
			return bestStart, bestGain
		}
		if gain <= bestGain {
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
		bestGain := 0
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
		if seg := NewMaskedSegment(WithBuffer(in[bestStart:bestEnd+1]), bestStart); seg.GetCompressionGains() > 0 {
			list.AppendValue(seg)
			i = bestEnd
		}
	}
	return list, FillSegmentGaps(in, list)
}
