package segments

import (
	"math"

	"github.com/sonalys/gompressor/compression"
	ll "github.com/sonalys/gompressor/linkedlist"
)

func getMaskedGain(in []byte, pos int) (int, bool) {
	var mask byte
	var notMask byte
	for i := 0; i < len(in); i++ {
		mask |= in[i]
		notMask |= ^in[i]
	}
	sizeMask := compression.Count1Bits(mask)
	sizeNotmask := compression.Count1Bits(notMask)
	if sizeMask+sizeNotmask == 16 {
		return math.MinInt, false
	}
	if sizeNotmask < sizeMask {
		mask = notMask
	}
	originalSize := len(in)
	compressedSize := calculateMaskedCompressedSize(mask, originalSize, pos)
	return originalSize - compressedSize, true
}

type windowInfo struct {
	pos  int
	gain int
}

func getBestStart(in []byte, cur *ll.ListEntry[*windowInfo], bestStart, bestEnd, bestGain, windowSize int) int {
	for {
		if cur.Prev == nil || cur.Prev.Value.gain < 0 {
			break
		}
		cur = cur.Prev
		prevStart := bestStart - windowSize/2
		if prevStart < 0 {
			break
		}
		gain, ok := getMaskedGain(in[prevStart:bestEnd], prevStart)
		if !ok {
			break
		}
		if gain < bestGain {
			break
		}
		bestStart = prevStart
		bestGain = gain
	}
	return bestStart
}

func getBestEnd(in []byte, cur *ll.ListEntry[*windowInfo], bestStart, bestEnd, bestGain, windowSize int) int {
	inLen := len(in)
	for {
		if cur.Next == nil || cur.Next.Value.gain < 0 {
			break
		}
		cur = cur.Next
		nextEnd := bestEnd + windowSize/2
		if nextEnd > inLen {
			nextEnd = inLen
		}
		gain, ok := getMaskedGain(in[bestStart:nextEnd], bestStart)
		if !ok {
			break
		}
		if gain < bestGain {
			break
		}
		bestEnd = nextEnd
	}
	return bestEnd
}

func CreateMaskedSegments(in []byte) (*ll.LinkedList[Segment], []byte) {
	list := &ll.LinkedList[Segment]{}
	inLen := len(in)
	const windowSize = 7
	gainList := *ll.NewLinkedList[*windowInfo]()
	for i := 0; i < inLen; i += windowSize / 2 {
		end := i + windowSize
		if end > inLen {
			end = inLen
		}
		gain, _ := getMaskedGain(in[i:end], i)
		entry := &windowInfo{
			pos:  i,
			gain: gain,
		}
		gainList.AppendValue(entry)
	}
	cur := gainList.Head
	for {
		if cur == nil {
			break
		}
		if cur.Value.gain < 0 {
			cur = cur.Next
			continue
		}
		entry := cur.Value
		bestGain := entry.gain
		bestStart := entry.pos
		bestEnd := bestStart + windowSize
		bestStart = getBestStart(in, cur, bestStart, bestEnd, bestGain, windowSize)
		bestEnd = getBestEnd(in, cur, bestStart, bestEnd, bestGain, windowSize)
		if seg := NewMaskedSegment(WithBuffer(in[bestStart:bestEnd]), bestStart); seg.GetCompressionGains() > 0 {
			list.AppendValue(seg)
			for {
				cur = cur.Next
				if cur == nil {
					break
				}
				if cur.Value.pos > bestEnd {
					break
				}
			}
			if cur == nil {
				break
			}
			cur.Prev = nil
			continue
		}
		cur = cur.Next
	}
	return list, FillSegmentGaps(in, list)
}
