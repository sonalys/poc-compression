package gompressor

import "bytes"

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

// getOtherRepeatingPos finds all other positions from the same byte value that also repeat the group found.
func getOtherRepeatingPos(
	posList []int64,
	buf, cmp []byte,
	conflictChecker map[int64]struct{},
	startOffset, endOffset int64,
) (resp, unusedPos []int64, conflict bool) {
	bufLen := int64(len(buf))
	for _, pos := range posList {
		startPos := pos - startOffset
		// We need to be sure no other char is growing the same repetition group as we are.
		// No matter which pos you start, once the group has grown, it will always have the same startPos.
		if _, conflict := conflictChecker[startPos]; conflict {
			return resp, unusedPos, true
		}
		endPos := pos + endOffset
		if startPos > bufLen || endPos > bufLen {
			break
		}
		if bytes.Equal(buf[startPos:endPos], cmp) {
			resp = append(resp, startPos)
			continue
		}
		unusedPos = append(unusedPos, pos)
	}
	return
}

// findStartSearchIndex finds the next position that is not contained by the current matching group.
func findStartSearchIndex(posList []int64, first int, endPos int64) int {
	startSearchIndex := first
	for {
		if posList[startSearchIndex] > endPos {
			break
		}
		startSearchIndex += 1
	}
	return startSearchIndex
}

// CreateRepeatingSegments detects repeating groups of bytes and index them into the same segment.
// It does it so by indexing the positions of each byte from 0..255, then using their positions for detecting
// repeating groups, since if they repeat, they should contain the same byte.
// We then grow each group as large as possible and index them.
func CreateRepeatingSegments(buf []byte) *LinkedList[Segment] {
	bufLen := int64(len(buf))
	byteMap := MapBytePos(buf)
	minSize := int64(4)
	conflictChecker := make(map[int64]struct{}, 1024)
	list := NewLinkedList[Segment]()
	for _, posList := range byteMap {
		for firstIdx, lastIdx := 0, len(posList)-1; lastIdx > firstIdx && posList[lastIdx]-posList[firstIdx] >= minSize; firstIdx, lastIdx = firstIdx+1, lastIdx-1 {
			startOffset := getStartOffset(buf, posList[firstIdx], posList[lastIdx])
			endOffset := getEndOffset(buf, posList[firstIdx], posList[lastIdx])
			// Ensure the repetition found is bigger than minSize.
			if endOffset+startOffset < minSize {
				continue
			}
			start := posList[firstIdx] - startOffset
			end := posList[firstIdx] + endOffset
			startSearchIndex := findStartSearchIndex(posList, firstIdx+1, end)

			segPos, unusedPos, conflict := getOtherRepeatingPos(posList[startSearchIndex:], buf, buf[start:end], conflictChecker, startOffset, endOffset)
			if conflict {
				continue
			}
			posList = unusedPos
			cur := NewSegment(TypeRepeatingGroup, start, buf[start:end])
			cur.AppendPos(segPos)
			for _, pos := range cur.Pos {
				conflictChecker[pos] = struct{}{}
			}
			list.AppendValue(cur)
			// Update posList to contain only positions that weren't matched yet.
			posList = unusedPos
			firstIdx, lastIdx = -1, len(posList)
		}
	}
	var prev int64
	// Reconstrut the uncompressed segments that connects all the segments we created.
	// We sort by POS here because they should be linear and crescent by pos.
	for _, seg := range sortAndFilterSegments(list, false) {
		// Prevent segment interpolation by removing the group on the pos.
		if prev >= seg.Pos {
			seg.RemovePos(seg.Pos)
			continue
		}
		list.AppendValue(NewSegment(TypeUncompressed, prev, buf[prev:seg.Pos]))
		prev = seg.Pos + int64(len(seg.Buffer))
	}
	// If there is any remaining uncompressed buffer after our last segment, we need to put it on the buffer as well.
	if prev < bufLen {
		list.AppendValue(NewSegment(TypeUncompressed, prev, buf[prev:]))
	}
	return list
}
