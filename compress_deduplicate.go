package gompressor

import "bytes"

// Deduplicate will find segments that are identical, besides position, and merge them.
func (s *Segment) Deduplicate() *Segment {
	s.ForEach(func(cur *Segment) {
		cur.Next.ForEach(func(iter *Segment) {
			if !bytes.Equal(cur.Buffer, iter.Buffer) || cur.Repeat != iter.Repeat || cur.Type != iter.Type {
				return
			}
			// if pos doesn't overflow, we continue with the merge operation.
			if _, err := cur.AddPos(iter.Pos); err == nil {
				next := iter.Remove()
				if iter == s {
					s = next
				}
			}
		})
	})
	return s
}
