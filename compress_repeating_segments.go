package gompressor

import "bytes"

func CreateRepeatingSegments(buf []byte) *Segment {
	bufLen := uint32(len(buf))
	byteMap := MapBytePos(buf)
	minSize := uint32(6)
	conflictChecker := make(map[uint32]struct{}, len(buf))
	head := &Segment{}
	cur := head
	for _, posList := range byteMap {
	nextPos:
		for curIndex, nextIndex := 0, 1; nextIndex < len(posList); curIndex, nextIndex = curIndex+1, nextIndex+1 {
			// Finds the next pos that is far away enough to make a minSize group.
			for {
				if nextIndex >= len(posList) {
					continue nextPos
				}
				if posList[nextIndex]-posList[curIndex] >= minSize {
					break
				}
				nextIndex++
			}
			curPos := posList[curIndex]
			nextPos := posList[nextIndex]
			var startOffset, endOffset uint32
			// Search for smallest startOffset in which both groups are still equal.
			for newStart := startOffset + 1; curPos >= newStart; newStart++ {
				if buf[curPos-newStart] == buf[nextPos-newStart] {
					startOffset = newStart
					continue
				}
				break
			}
			// Search for the biggest endOffset in which both groups are still equal.
			for newEnd := endOffset + 1; nextPos+newEnd < bufLen; newEnd++ {
				if buf[curPos+newEnd] == buf[nextPos+newEnd] {
					endOffset = newEnd
					continue
				}
				break
			}
			// Ensure the repetition found is bigger than minSize.
			if endOffset+startOffset < minSize {
				continue
			}
			startPos := curPos - startOffset
			endPos := curPos + endOffset
			groupPos := []uint32{startPos, nextPos - startOffset}
			newPosList := make([]uint32, 0, len(posList))
			cmpBuf := buf[startPos:endPos]
			for _, pos := range posList[nextIndex+1:] {
				startPos := pos - startOffset
				// We need to be sure no other char is growing the same repetition group as we are.
				// No matter which pos you start, once the group has grown, it will always have the same startPos.
				if _, conflict := conflictChecker[startPos]; conflict {
					continue nextPos
				}
				endPos := pos + endOffset
				if startPos > bufLen || endPos > bufLen {
					break
				}
				if bytes.Equal(buf[startPos:endPos], cmpBuf) {
					groupPos = append(groupPos, startPos)
					continue
				}
				newPosList = append(newPosList, pos)
			}
			seg := &Segment{
				Type:   TypeRepeatingGroup,
				Repeat: 1,
				Buffer: buf[startPos:endPos],
				Pos:    groupPos,
			}
			if seg.GetCompressionGains() > 0 {
				for _, pos := range seg.Pos {
					conflictChecker[pos] = struct{}{}
				}
				cur = cur.Append(seg)
				// Update newPosList with positions that are still not used in any repetition group.
				posList = newPosList
				curIndex, nextIndex = -1, 0
			}
		}
	}
	// Remove empty head.
	head = head.Remove()
	var prev uint32
	for _, seg := range sortAndFilterSegments(head) {
		// Prevent segment interpolation by removing the group on the pos.
		if prev > seg.Pos {
			seg.RemovePos(seg.Pos)
			continue
		}
		cur = cur.Append(NewSegment(TypeUncompressed, prev, 1, buf[prev:seg.Pos]))
		prev = seg.Pos + uint32(len(seg.Buffer))
	}
	if prev < bufLen {
		cur.Append(NewSegment(TypeUncompressed, prev, 1, buf[prev:]))
	}
	return head
}
