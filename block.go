package gompressor

import (
	"bytes"
	"sort"
)

type Block struct {
	Size   uint32
	Head   *Segment
	Buffer []byte
}

// Remove dereferences this segment from the linked list.
func (b *Block) Remove(s *Segment) {
	if s.Previous == nil {
		b.Head = s.Next
	} else {
		s.Previous.Next = s.Next
	}
	if s.Next != nil {
		s.Next.Previous = s.Previous
	}
}

// Deduplicate will find segments that are identical, besides position, and merge them.
func (b *Block) Deduplicate() {
	b.Head.ForEach(func(cur *Segment) {
		cur.Next.ForEach(func(iter *Segment) {
			if !bytes.Equal(cur.Buffer, iter.Buffer) || cur.Repeat != iter.Repeat || cur.Type != iter.Type {
				return
			}
			// if pos doesn't overflow, we continue with the merge operation.
			if _, err := cur.AddPos(iter.Pos); err == nil {
				b.Remove(iter)
			}
		})
	})
}

// Optimize is responsible for finding segments that are causing byte compression gain to be negative, and try to
// revert it.
func (b *Block) Optimize() {
	// For making the logic easier on the POC, we just use this slice to sort by position.
	orderedSegments := make([]SegmentPosMap, 0, b.Size)
	// If we are not gaining any delta size, we just move it to the uncompressed buffer.
	b.Head.ForEach(func(cur *Segment) {
		if cur.GetCompressionGains() > 0 {
			return
		}
		for _, pos := range cur.Pos {
			orderedSegments = append(orderedSegments, SegmentPosMap{
				Pos:     pos,
				Segment: cur,
			})
		}
		b.Remove(cur)
	})
	sort.Slice(orderedSegments, func(i, j int) bool {
		return orderedSegments[i].Pos < orderedSegments[j].Pos
	})
	for _, entry := range orderedSegments {
		cur, pos := entry.Segment, entry.Pos
		segBuf := cur.Decompress()
		bufLen := uint32(len(b.Buffer))
		if pos < bufLen {
			panic("buffer optimization should be linear")
		}
		b.Buffer = append(b.Buffer, segBuf...)
	}
}

func (s *Segment) ForEach(f func(*Segment)) {
	cur := s
	for {
		if cur == nil {
			break
		}
		f(cur)
		cur = cur.Next
	}
}
