package segments

import (
	"math"
	"sync"

	ll "github.com/sonalys/gompressor/linkedlist"
)

func getGain(matched []int, size int) int {
	matchedLen := len(matched)
	maxPos := matched[len(matched)-1]
	return matchedLen*size - calculateGroupCompressedSize(matchedLen, size, maxPos)
}

func growIterator(v int) int {
	if v > 0 {
		return v + 1
	}
	return v - 1
}

func abs(v int) int {
	return (v>>63 | 1) * v
}

var pool = sync.Pool{
	New: func() any {
		list := make([]int, 0, 100)
		return &list
	},
}

func GrowOffset(in []byte, collision []bool, posList []int, offset, prevOffset int) ([]int, int) {
	// Copy, so we are able to return original posList in case nothing matches.
	localPosList := posList
	var bestMatched []int
	var bestOffset int
	var bestGain int = math.MinInt
	// We keep 3 different levels of records here,
	// Best level, the best gain we could achieve, for which offset and which positions.
	// Offset level, the best gain and positions for all possible start positions.
	// Start level, the best positions and gain given a fixed start position.
	for offset := offset; ; offset = growIterator(offset) {
		size := abs(offset-prevOffset) + 1
		var offsetGain int = math.MinInt
		var offsetMatched []int
		for i := 0; i < len(localPosList)-1; i++ {
			curPos := localPosList[i]
			if collision[curPos] {
				// No need to process this pos for any offset.
				posList = posList[i+1:]
				localPosList = posList
				continue
			}
			if curPos+offset < 0 {
				continue
			}
			// get pre-alloc slice from sync.pool and empty it.
			startMatched := (*pool.Get().(*[]int))[:0]
			startMatched = append(startMatched, curPos)
			lastMatchedPos := curPos
			for j := i + 1; j < len(localPosList); j++ {
				nextPos := localPosList[j]
				if collision[nextPos] || nextPos+offset >= len(in) {
					continue
				}
				if nextPos-lastMatchedPos <= size-1 {
					continue
				}
				if in[curPos+offset] == in[nextPos+offset] {
					lastMatchedPos = nextPos
					startMatched = append(startMatched, nextPos)
				}
			}
			if len(startMatched) < 2 {
				continue
			}
			if gain := getGain(startMatched, size); gain > offsetGain {
				offsetGain = gain
				offsetMatched = startMatched
				continue
			}
			pool.Put(&startMatched)
		}
		if offsetGain <= bestGain {
			break
		}
		bestGain = offsetGain
		bestMatched = offsetMatched
		bestOffset = offset
		localPosList = offsetMatched
	}
	if bestOffset == 0 {
		return posList, 0
	}
	return bestMatched, bestOffset
}

func detectOffsetCollision(collision []bool, pos, size int) bool {
	for offset := 0; offset < size; offset++ {
		if collision[pos+offset] {
			return true
		}
	}
	return false
}

func registerBytes(collision []bool, posList []int, size int) {
	for _, pos := range posList {
		for offset := 0; offset < size; offset++ {
			collision[pos+offset] = true
		}
	}
}

func appendUncollidedPos(posList []int, collision []bool, seg *SegmentGroup, startOffset, size int) {
	for i := 0; i < len(posList); i++ {
		pos := posList[i] + startOffset
		if detectOffsetCollision(collision, pos, size) {
			continue
		}
		seg.appendPos(pos)
	}
}

// CreateGroupSegments should linearly detect repeating groups, without overlapping.
func CreateGroupSegments(in []byte) (*ll.LinkedList[Segment], []byte) {
	bufLen := len(in)
	byteMap := MapBytePos(in)
	bytePop := GetBytePopularity(byteMap)
	collision := make([]bool, bufLen)
	list := ll.NewLinkedList[Segment]()
	// We start by searching groups by the less frequent bytes.
	for _, char := range bytePop {
		posList := byteMap[char]
		if len(posList) == 0 {
			continue
		}
		var startOffset, endOffset int
		posList, startOffset = GrowOffset(in, collision, posList, -1, 0)
		posList, endOffset = GrowOffset(in, collision, posList, 1, startOffset)
		size := endOffset - startOffset + 1
		if size < 2 || len(posList) < 2 {
			continue
		}
		startPos := posList[0] + startOffset
		endPos := posList[0] + endOffset + 1
		segment := NewGroupSegment(in[startPos:endPos])
		appendUncollidedPos(posList, collision, segment, startOffset, size)
		if segment.GetCompressionGains() <= 0 {
			continue
		}
		registerBytes(collision, segment.pos, size)
		list.AppendValue(segment)
	}
	return list, FillSegmentGaps(in, list)
}
