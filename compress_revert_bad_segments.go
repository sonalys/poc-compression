package gompressor

// RevertBadSegments is responsible for reverting bad segments.
func RevertBadSegments(buf []byte, list *LinkedList[*Segment], size int) []byte {
	orderedSegments := sortAndFilterSegments(list, false, func(cur *ListEntry[*Segment]) bool {
		if cur.Value.GetCompressionGains() <= 0 {
			cur.Remove()
			return true
		}
		return false
	})
	for _, entry := range orderedSegments {
		cur, pos := entry.Segment, entry.Pos
		if pos < int(len(buf)) {
			panic("decompression should be linear")
		}
		buf = append(buf, cur.Decompress()...)
	}
	return buf
}
