package gompressor

// RevertBadSegments is responsible for reverting bad segments.
func (s *Segment) RevertBadSegments(size uint32) (*Segment, []byte) {
	orderedSegments := sortAndFilterSegments(s, func(cur *Segment) bool {
		if cur.GetCompressionGains() <= 0 {
			next := cur.Remove()
			if s == cur {
				s = next
			}
			return true
		}
		return false
	})
	out := make([]byte, 0, size)
	for _, entry := range orderedSegments {
		cur, pos := entry.Segment, entry.Pos
		bufLen := uint32(len(out))
		if pos < bufLen {
			panic("reconstruction should be linear")
		}
		out = append(out, cur.Decompress()...)
	}
	return s, out
}
