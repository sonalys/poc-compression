package gompressor

// RevertBadSegments is responsible for reverting bad segments.
func RevertBadSegments[S BlockSize](list *LinkedList[Segment[S]], size S) []byte {
	orderedSegments := sortAndFilterSegments(list, false, func(cur *ListEntry[Segment[S]]) bool {
		if cur.Value.GetCompressionGains() <= 0 {
			cur.Remove()
			return true
		}
		return false
	})
	out := make([]byte, 0, size)
	for _, entry := range orderedSegments {
		cur, pos := entry.Segment, entry.Pos
		bufLen := S(len(out))
		if pos < bufLen {
			panic("reconstruction should be linear")
		}
		out = append(out, cur.Decompress()...)
	}
	return out
}
