package gompressor

import "bytes"

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

func getNextIndex(posList []int64, minSize int64, curIndex int, nextIndex int) (int, bool) {
	for {
		if nextIndex >= len(posList) {
			return -1, false
		}
		if posList[nextIndex]-posList[curIndex] >= minSize {
			break
		}
		nextIndex++
	}
	return nextIndex, true
}

func getOtherRepeatingPos(
	posList []int64,
	bufLen int64,
	buf, cmp []byte,
	conflictChecker map[int64]struct{},
	startOffset, endOffset int64,
) (resp, unusedPos []int64, conflict bool) {
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

func CreateRepeatingSegments(buf []byte) *LinkedList[Segment] {
	bufLen := int64(len(buf))
	byteMap := MapBytePos(buf)
	minSize := int64(4)
	conflictChecker := make(map[int64]struct{}, 1024)
	list := NewLinkedList[Segment]()
	for _, posList := range byteMap {
		// TODO: change this strategy to use first and last value, and they meet in the middle.
		// this will allow us to detect repeating groups that contains cur and next, and skip positions that are contained in the group.
		for curIndex, nextIndex := 0, 1; nextIndex < len(posList); curIndex, nextIndex = curIndex+1, nextIndex+1 {
			// Finds the next pos that is far away enough to make a minSize group.
			nextIndex, ok := getNextIndex(posList, minSize, curIndex, nextIndex)
			if !ok {
				continue
			}
			startOffset := getStartOffset(buf, posList[curIndex], posList[nextIndex])
			endOffset := getEndOffset(buf, posList[curIndex], posList[nextIndex])
			// Ensure the repetition found is bigger than minSize.
			if endOffset+startOffset < minSize {
				continue
			}
			start := posList[curIndex] - startOffset
			end := posList[curIndex] + endOffset

			pos, unused, conflict := getOtherRepeatingPos(posList[nextIndex+1:], bufLen, buf, buf[start:end], conflictChecker, startOffset, endOffset)
			if conflict {
				continue
			}
			pos = append(pos, start, posList[nextIndex]-startOffset)
			posList = unused
			cur := &Segment{
				Type:   TypeRepeatingGroup,
				Repeat: 1,
				Buffer: buf[start:end],
			}
			cur.AppendPos(pos)
			for _, pos := range cur.Pos {
				conflictChecker[pos] = struct{}{}
			}
			list.AppendValue(cur)
			// Update newPosList with positions that are still not used in any repetition group.
			posList = unused
			curIndex, nextIndex = -1, 0
		}
	}
	var prev int64
	for _, seg := range sortAndFilterSegments(list, false) {
		// Prevent segment interpolation by removing the group on the pos.
		if prev >= seg.Pos {
			seg.RemovePos(seg.Pos)
			continue
		}
		list.AppendValue(NewSegment(TypeUncompressed, prev, 1, buf[prev:seg.Pos]))
		prev = seg.Pos + int64(len(seg.Buffer))
	}
	if prev < bufLen {
		list.AppendValue(NewSegment(TypeUncompressed, prev, 1, buf[prev:]))
	}
	return list
}
