package segments

import (
	"math"

	"github.com/sonalys/gompressor/compression"
	ll "github.com/sonalys/gompressor/linkedlist"
)

func recalculateWindowMaskGain(in []byte, enableInvert bool, mask byte, pos, size int) (newMask byte, gain int, invert, ok bool) {
	mask, shouldEnableInvert, _, maskSize := compression.MaskRegisterByte(mask, in[pos])
	enableInvert = enableInvert || shouldEnableInvert
	if enableInvert {
		maskSize++
	}
	if maskSize == 8 {
		return mask, math.MinInt, enableInvert, false
	}
	compressedSize := (size*maskSize + 8 - 1) / 8
	return mask, size - compressedSize, enableInvert, true
}

func getBestEnd(in []byte, enableInvert bool, start, bestEnd, bestGain int) (int, int, bool) {
	var mask byte
	var gain, maskSize int
	var ok, curInvert bool
	mask, maskSize, curInvert, _ = compression.MaskRegisterBuffer(in[start : bestEnd+1])
	enableInvert = enableInvert || curInvert
	if enableInvert {
		maskSize++
	}
	if maskSize == 8 {
		return bestEnd, gain, enableInvert
	}
	for newEnd := bestEnd + 1; bestEnd < len(in)-1; newEnd++ {
		mask, gain, curInvert, ok = recalculateWindowMaskGain(in, enableInvert, mask, newEnd, newEnd-start)
		if !ok || gain < bestGain {
			break
		}
		enableInvert = enableInvert || curInvert
		bestEnd = newEnd
		bestGain = gain
	}
	return bestEnd, bestGain, enableInvert
}

func getBestStart(in []byte, enableInvert bool, bestStart, end, bestGain int, minSize int) (int, int, bool) {
	var gain, maskSize int
	var ok, curInvert bool
	for newStart := bestStart + 1; bestStart+minSize < end; newStart++ {
		var mask byte
		// We are removing the first byte, so we always have to recalculate the masks.
		for i := newStart; i <= end; i++ {
			mask, enableInvert, _, maskSize = compression.MaskRegisterByte(mask, in[i])
			if maskSize == 8 || enableInvert && maskSize == 7 {
				return bestStart, bestGain, enableInvert
			}
		}
		_, gain, curInvert, ok = recalculateWindowMaskGain(in, enableInvert, mask, newStart, end-newStart)
		if !ok {
			break
		}
		if gain < bestGain {
			break
		}
		enableInvert = enableInvert || curInvert
		bestStart = newStart
		bestGain = gain
	}
	return bestStart, bestGain, enableInvert
}

func CreateMaskedSegments(in []byte) (*ll.LinkedList[Segment], []byte) {
	list := &ll.LinkedList[Segment]{}
	inLen := len(in)
	const minSize = 6
	for i := 0; i < inLen-minSize-1; i++ {
		bestGain := math.MinInt
		bestStart := i
		bestEnd := i
		enableInvert := false
		if bestEnd > inLen {
			bestEnd = inLen
		}
		prevBestGain := bestGain
		for {
			if newEnd, gain, invert := getBestEnd(in, enableInvert, bestStart, bestEnd, bestGain); gain > bestGain {
				bestEnd = newEnd
				bestGain = gain
				enableInvert = enableInvert || invert
			}
			if newStart, gain, invert := getBestStart(in, enableInvert, bestStart, bestEnd, bestGain, minSize); gain > bestGain {
				bestStart = newStart
				bestGain = gain
				enableInvert = enableInvert || invert
			}
			if prevBestGain == bestGain {
				break
			}
			prevBestGain = bestGain
		}
		// println(bestStart, bestEnd, bestGain)
		if seg := NewMaskedSegment(in[bestStart:bestEnd+1], bestStart); seg.GetCompressionGains() > 0 {
			list.AppendValue(seg)
			i = bestEnd
		}
	}
	return list, FillSegmentGaps(in, list)
}
