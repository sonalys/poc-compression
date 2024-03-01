package segments

import (
	"math"

	"github.com/sonalys/gompressor/compression"
	ll "github.com/sonalys/gompressor/linkedlist"
)

func recalculateWindowMaskGain(b byte, enableInvert bool, mask byte, size int) (newMask byte, gain int, invert, ok bool) {
	mask, shouldEnableInvert, _, maskSize := compression.MaskRegisterByte(mask, b)
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

func getBestEnd(in []byte, masks []*maskCalculator, start, bestEnd, bestGain int) (int, int) {
	for newEnd := bestEnd + 1; newEnd < len(in)-1; newEnd++ {
		curBestGain := bestGain
		for i := range masks {
			gain, _ := masks[i].registerByte(in[newEnd], newEnd-start, newEnd-start)
			if gain < curBestGain {
				break
			}
			bestEnd = newEnd
			curBestGain = gain
		}
		if curBestGain < bestGain {
			break
		}
		bestGain = curBestGain
	}
	return bestEnd, bestGain
}

func getBestStart(in []byte, masks []*maskCalculator, bestStart, end, bestGain int, minSize int) (int, int) {
	var ok bool
	for newStart := bestStart + 1; bestStart+minSize < end; newStart++ {
		curBestGain := bestGain
		// We are removing the first byte, so we always have to recalculate the masks.
		for i := newStart; i <= end; i++ {
			for _, mask := range masks {
				curBestGain, ok = mask.registerByte(in[i], end-bestStart, i-newStart)
				if !ok {
					break
				}
			}
		}
		if curBestGain < bestGain {
			break
		}
		bestStart = newStart
		bestGain = curBestGain
	}
	return bestStart, bestGain
}

type mask struct {
	mask         byte
	enableInvert bool
}

type maskCalculator struct {
	masks []mask
	gain  []int
}

func newMaskCalculator(maskCount int) *maskCalculator {
	resp := &maskCalculator{
		masks: make([]mask, maskCount),
		gain:  make([]int, maskCount),
	}
	return resp
}

func (m *maskCalculator) calcGain() (gain int) {
	for i := range m.gain {
		gain += m.gain[i]
	}
	return
}

func (m *maskCalculator) registerByte(b byte, size, pos int) (gain int, ok bool) {
	if len(m.masks) == 0 {
		panic("unitialized mask")
	}
	idx := pos % len(m.masks)
	mask := &m.masks[idx]
	mask.mask, gain, mask.enableInvert, ok = recalculateWindowMaskGain(b, mask.enableInvert, mask.mask, size)
	m.gain[idx] = gain
	return m.calcGain(), ok
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
		var masks []*maskCalculator
		for j := 1; j < 4; j++ {
			masks = append(masks, newMaskCalculator(j))
		}
		prevBestGain := bestGain
		for {
			if newEnd, gain := getBestEnd(in, masks, bestStart, bestEnd, bestGain); gain > bestGain {
				bestEnd = newEnd
				bestGain = gain
			}
			if newStart, gain := getBestStart(in, masks, bestStart, bestEnd, bestGain, minSize); gain > bestGain {
				bestStart = newStart
				bestGain = gain
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
