package gompressor

import "math"

func getGain(matched []int, size int) int {
	matchedLen := len(matched)
	maxPos := matched[len(matched)-1]
	return matchedLen*size - GetCompressedSize(TypeRepeatingGroup, 1, maxPos, matchedLen, size)
}

func growIterator(v int) int {
	if v > 0 {
		return v + 1
	}
	return v - 1
}

func cmp(buf []byte, pos, next, offset int) bool {
	bufLen := len(buf)
	if offset > 0 {
		if next+offset >= bufLen {
			return false
		}
	} else if pos+offset < 0 {
		return false
	}
	return buf[pos+offset] == buf[next+offset]
}

func abs(v int) int {
	return (v>>63 | 1) * v
}

func GrowOffset(buf []byte, collision []bool, posList []int, offset, prevSize int) ([]int, int, int) {
	var bestMatched []int
	var bestOffset int
	var bestGain int = math.MinInt
	// We keep 3 different levels of records here,
	// Best level, the best gain we could achieve, for which offset and which positions.
	// Offset level, the best gain and positions for all possible start positions.
	// Start level, the best positions and gain given a fixed start position.
	for offset := offset; ; offset = growIterator(offset) {
		var offsetGain int = math.MinInt
		var offsetMatched []int
		absOffset := abs(offset)
		for i := 0; i < len(posList); i++ {
			curPos := posList[i]
			if collision[curPos] {
				continue
			}
			startMatched := make([]int, 0, 100)
			startMatched = append(startMatched, curPos)
			for j := i + 1; j < len(posList); j++ {
				nextPos := posList[j]
				if collision[nextPos] || nextPos-curPos < absOffset {
					continue
				}
				if cmp(buf, curPos, nextPos, offset) {
					startMatched = append(startMatched, nextPos)
				}
			}
			if len(startMatched) < 2 {
				continue
			}
			if gain := getGain(startMatched, prevSize+absOffset); gain > offsetGain {
				offsetGain = gain
				offsetMatched = startMatched
			}
		}
		if offsetGain <= bestGain {
			break
		}
		bestGain = offsetGain
		bestMatched = offsetMatched
		bestOffset = offset
		posList = offsetMatched
	}
	if bestGain <= 0 {
		return posList, 0, 0
	}
	return bestMatched, bestOffset, bestGain
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

func appendUncollidedPos(posList []int, collision []bool, seg *Segment, startOffset, size int) {
	for i := 0; i < len(posList); i++ {
		pos := posList[i] + startOffset
		if detectOffsetCollision(collision, pos, size) {
			continue
		}
		if i == 0 || posList[i]-posList[i-1] > size {
			seg.AppendPos(pos)
		}
	}
}

// CreateRepeatingSegments should linearly detect repeating groups, without overlapping.
func CreateRepeatingSegments(buf []byte) (*LinkedList[*Segment], []byte) {
	bufLen := len(buf)
	minSize := 3
	byteMap := MapBytePos(buf)
	bytePop := GetBytePopularity(byteMap)
	collision := make([]bool, bufLen)
	list := NewLinkedList[*Segment]()
	var charCount int
	// We start by searching groups by the less frequent bytes.
	for _, char := range bytePop {
		posList := byteMap[char]
		if len(posList) == 0 {
			continue
		}
		var startOffset, endOffset int
		var startGain, endGain int
		var bestStartList, bestEndList []int
		bestStartList, startOffset, startGain = GrowOffset(buf, collision, posList, -1, 1)
		bestEndList, endOffset, endGain = GrowOffset(buf, collision, posList, 1, 1)
		if startGain > endGain {
			posList, endOffset, _ = GrowOffset(buf, collision, bestStartList, 1, abs(startOffset)+1)
		} else {
			posList, startOffset, _ = GrowOffset(buf, collision, bestEndList, -1, abs(endOffset)+1)
		}

		size := endOffset - startOffset
		if size < minSize || len(posList) < 2 {
			continue
		}
		startPos := posList[0] + startOffset
		endPos := posList[0] + endOffset
		seg := NewSegment(TypeRepeatingGroup, buf[startPos:endPos])
		appendUncollidedPos(posList, collision, seg, startOffset, size)
		if seg.GetCompressionGains() <= 0 {
			continue
		}
		charCount += seg.GetCompressionGains()
		// verifySegment(buf, seg)
		registerBytes(collision, seg.Pos, size)
		list.AppendValue(seg)
	}
	return list, FillSegmentGaps(buf, list)
}
