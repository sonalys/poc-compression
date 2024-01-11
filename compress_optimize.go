package gompressor

// RemoveNegativeSegments is responsible for finding segments that are causing byte compression gain to be negative, and try to
// revert it.
func (s *Segment) RemoveNegativeSegments(size uint32) []byte {
	orderedSegments := sortAndFilterSegments(s, func(cur *Segment) bool {
		return cur.GetCompressionGains() <= 0
	})
	out := make([]byte, 0, size)
	for _, entry := range orderedSegments {
		cur, pos := entry.Segment, entry.Pos
		bufLen := uint32(len(out))
		if pos < bufLen {
			panic("reconstruction should be linear")
		}
		out = append(out, cur.Decompress()...)
		cur.Remove()
	}
	return out
}
