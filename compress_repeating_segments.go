package gompressor

import "bytes"

func getStartOffset[S BlockSize](buf []byte, curPos, nextPos S) S {
	var startOffset S
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

func getEndOffset[S BlockSize](buf []byte, curPos, nextPos S) S {
	var endOffset S
	bufLen := S(len(buf))
	for newEnd := endOffset + 1; nextPos+newEnd < bufLen; newEnd++ {
		if buf[curPos+newEnd] == buf[nextPos+newEnd] {
			endOffset = newEnd
			continue
		}
		break
	}
	return endOffset
}

func getNextIndex[S BlockSize](posList []S, minSize S, curIndex int, nextIndex int) (int, bool) {
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

func getOtherRepeatingPos[S BlockSize](
	posList []S,
	bufLen S,
	buf, cmp []byte,
	conflictChecker map[S]struct{},
	startOffset, endOffset S,
) (resp, unusedPos []S, conflict bool) {
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

func CreateRepeatingSegments[S BlockSize](buf []byte) *LinkedList[Segment[S]] {
	bufLen := S(len(buf))
	byteMap := MapBytePos[S](buf)
	minSize := S(4)
	conflictChecker := make(map[S]struct{}, 1000)
	list := NewLinkedList[Segment[S]]()
	for _, posList := range byteMap {
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
			groupPos := []S{start, posList[nextIndex] - startOffset}

			pos, unused, conflict := getOtherRepeatingPos(posList[nextIndex+1:], bufLen, buf, buf[start:end], conflictChecker, startOffset, endOffset)
			if conflict {
				continue
			}
			posList = unused

			seg := &Segment[S]{
				Type:   TypeRepeatingGroup,
				Repeat: 1,
				Buffer: buf[start:end],
				Pos:    append(groupPos, pos...),
			}
			if seg.GetCompressionGains() > 0 {
				for _, pos := range seg.Pos {
					conflictChecker[pos] = struct{}{}
				}
				list.AppendValue(seg)
				// Update newPosList with positions that are still not used in any repetition group.
				posList = unused
				curIndex, nextIndex = -1, 0
			}
		}
	}
	var prev S
	for _, seg := range sortAndFilterSegments(list, false) {
		// Prevent segment interpolation by removing the group on the pos.
		if prev > seg.Pos {
			seg.RemovePos(seg.Pos)
			continue
		}
		list.AppendValue(NewSegment(TypeUncompressed, prev, 1, buf[prev:seg.Pos]))
		prev = seg.Pos + S(len(seg.Buffer))
	}
	if prev < bufLen {
		list.AppendValue(NewSegment(TypeUncompressed, prev, 1, buf[prev:]))
	}
	return list
}
