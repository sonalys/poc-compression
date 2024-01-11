package gompressor

import "bytes"

// Deduplicate will find segments that are identical, besides position, and merge them.
func (s *Segment) Deduplicate() {
	s.ForEach(func(cur *Segment) {
		cur.Next.ForEach(func(iter *Segment) {
			if !bytes.Equal(cur.Buffer, iter.Buffer) || cur.Repeat != iter.Repeat || cur.Type != iter.Type {
				return
			}
			// if pos doesn't overflow, we continue with the merge operation.
			if _, err := cur.AddPos(iter.Pos); err == nil {
				iter.Remove()
			}
		})
	})
}
