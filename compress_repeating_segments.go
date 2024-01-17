package gompressor

import (
	"bytes"
)

// getStartOffset grows a byte group from the start position, finding the largest group possible.
// this group must repeat in curPos and nextPos.
func getStartOffset(buf []byte, curPos, nextPos int64) int64 {
	var startOffset int64
	// Search for smallest startOffset in which both groups are still equal.
	for newStart := startOffset + 1; curPos >= newStart; newStart++ {
		if buf[curPos-newStart] == buf[nextPos-newStart] {
			startOffset = newStart
			continue
		}
		break
	}
	return startOffset
}

// getEndOffset grows a byte group from the end position, finding the largest group possible.
// this group must repeat in curPos and nextPos.
func getEndOffset(buf []byte, curPos, nextPos int64) int64 {
	var endOffset int64
	bufLen := int64(len(buf))
	for newEnd := endOffset + 1; nextPos+newEnd < bufLen; newEnd++ {
		if buf[curPos+newEnd] == buf[nextPos+newEnd] {
			endOffset = newEnd
			continue
		}
		break
	}
	return endOffset
}

// matchRepeatingGroup finds all other positions from the same byte value that also repeat the group found.
func matchRepeatingGroup(list []int64, buf, cmp []byte, startOffset, endOffset int64) (resp []int64) {
	bufLen := int64(len(buf))
	var lastPos int64 = -1
	for _, pos := range list {
		startPos := pos - startOffset
		endPos := pos + endOffset
		// Prevent start and end positions from panicking,
		// also prevents position intersection from previous position with the current.
		if startPos > bufLen || endPos > bufLen {
			break
		}
		if lastPos > -1 && lastPos+endOffset >= startPos {
			continue
		}
		if bytes.Equal(buf[startPos:endPos], cmp) {
			resp = append(resp, startPos)
			lastPos = startPos
			continue
		}
	}
	return
}

func findBytePosIndex(posList []int64, i int64) (int, bool) {
	for j, pos := range posList {
		if i == pos {
			return j, true
		}
	}
	return -1, false
}

// CreateRepeatingSegments2 should linearly detect repeating groups, without overlapping.
func CreateRepeatingSegments2(buf []byte) (*LinkedList[Segment], []byte) {
	minSize := int64(3)
	list := NewLinkedList[Segment]()
	byteMap := MapBytePos(buf)
	bufLen := int64(len(buf))
	for curPos := int64(0); curPos < bufLen; curPos++ {
		char := buf[curPos]
		bytePosList := byteMap[char]
		posIndex, ok := findBytePosIndex(bytePosList, curPos)
		if !ok {
			continue
		}
		searchPosList := bytePosList[posIndex+1:]
		for j := 0; j < len(searchPosList); j++ {
			nextPos := searchPosList[j]
			if _, ok := findBytePosIndex(bytePosList, nextPos); !ok {
				continue
			}
			startOffset := getStartOffset(buf, curPos, nextPos)
			endOffset := getEndOffset(buf, curPos, nextPos)
			// Prevent positions too close to form a group.
			// Also prevents forming a group that intersects it's positions.
			if endOffset-startOffset < minSize || endOffset-startOffset > nextPos-curPos {
				continue
			}
			groupStart := curPos - startOffset
			groupEnd := curPos + endOffset
			groupBuf := buf[groupStart:groupEnd]
			// repeating group of minSize found in curPos and nextPos.
			matched := matchRepeatingGroup(searchPosList[j:], buf, groupBuf, startOffset, endOffset)
			matched = append(matched, groupStart)
			nonCollidingPos := make([]int64, 0, len(matched))
			var bytesToRemove [256][]int64
		nextPos:
			for _, pos := range matched {
				// First check for byte collision with other segments.
				for k, char := range groupBuf {
					charPos := pos + int64(k)
					if _, ok := findBytePosIndex(byteMap[char], charPos); !ok {
						continue nextPos
					}
				}
				// No byte collisions means we can create our segment without any further issues.
				// So here register all these bytes to be removed from the indexing.
				for k, char := range groupBuf {
					charPos := pos + int64(k)
					bytesToRemove[char] = append(bytesToRemove[char], charPos)
				}
				// Position had no collisions, so it can be added to the segment.
				nonCollidingPos = append(nonCollidingPos, pos)
			}
			if len(nonCollidingPos) == 0 {
				continue
			}
			cur := NewSegment(TypeRepeatingGroup, groupBuf, nonCollidingPos...)
			// If the segment is net negative, it's not worth to add it.
			if cur.GetCompressionGains() <= 0 {
				continue
			}
			for char, posList := range bytesToRemove {
				for _, pos := range posList {
					posIndex, ok := findBytePosIndex(byteMap[char], pos)
					if !ok {
						panic("same byte mapped twice in the same segment")
					}
					byteMap[char] = append(byteMap[char][:posIndex], byteMap[char][posIndex+1:]...)
				}
			}
			// Adds the segment to the list and remove all the bytes used from indexing.
			list.AppendValue(cur)
			break
		}
	}
	return list, FillSegmentGaps(buf, list)
}

func RemoveConflicts(list *LinkedList[Segment]) {
	orderedSegments := sortAndFilterSegments(list, false)
	for i := 0; i < len(orderedSegments); i++ {
		cur := orderedSegments[i]
		endPos := cur.Pos + int64(len(cur.Buffer))*int64(cur.Repeat)
		conflicts := make([]SegmentPosMap, 0, 5)
		var conflictGains int64
		for j := i + 1; j < len(orderedSegments); j++ {
			next := orderedSegments[j]
			if next.Pos > endPos {
				break
			}
			conflicts = append(conflicts, next)
			conflictGains += next.GetCompressionGains()
		}
		if len(conflicts) == 0 {
			continue
		}
		if cur.GetCompressionGains() > conflictGains {
			for i := range conflicts {
				conflicts[i].RemovePos(conflicts[i].Pos)
			}
			orderedSegments = append(orderedSegments[:i+1], orderedSegments[1+i+len(conflicts):]...)
			continue
		}
		cur.RemovePos(cur.Pos)
	}
}
