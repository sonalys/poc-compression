package gompressor

// RevertBadSegments is responsible for reverting bad segments.
func RevertBadSegments(list *LinkedList[Segment], size int64) []byte {
	orderedSegments := sortAndFilterSegments(list, false, func(cur *ListEntry[Segment]) bool {
		if cur.Value.GetCompressionGains() <= 0 {
			cur.Remove()
			return true
		}
		return false
	})
	out := make([]byte, 0, size)
	for _, entry := range orderedSegments {
		cur, pos := entry.Segment, entry.Pos
		bufLen := int64(len(out))
		if pos < bufLen {
			panic("reconstruction should be linear")
		}
		out = append(out, cur.Decompress()...)
	}
	return out
}