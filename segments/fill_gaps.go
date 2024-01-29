package segments

import (
	"fmt"
	"sort"

	ll "github.com/sonalys/gompressor/linkedlist"
)

type SegmentPosMap[T Segment] struct {
	Pos     int
	Segment T
	Entry   *ll.ListEntry[T]
}

func SortAndFilterSegments[T Segment](list *ll.LinkedList[T], sortType bool, filters ...func(*ll.ListEntry[T]) bool) []SegmentPosMap[T] {
	out := make([]SegmentPosMap[T], 0, list.Len)
	list.ForEach(func(cur *ll.ListEntry[T]) {
		curValue := cur.Value
		if len(cur.Value.GetPos()) == 0 {
			return
		}
		for _, filter := range filters {
			if !filter(cur) {
				return
			}
		}
		for _, pos := range curValue.GetPos() {
			out = append(out, SegmentPosMap[T]{
				Pos:     pos,
				Entry:   cur,
				Segment: curValue,
			})
		}
	})
	sort.Slice(out, func(i, j int) bool {
		if sortType {
			t1, t2 := out[i].Segment.GetType(), out[j].Segment.GetType()
			// We layer the logic by segment type, so some segments should decompress first than others.
			if t1 != t2 {
				return t1 < t2
			}
		}
		return out[i].Pos < out[j].Pos
	})
	return out
}

func removeBadSegments[T Segment](entry *ll.ListEntry[T]) bool {
	if entry.Value.GetCompressionGains() <= 0 {
		entry.Remove()
		return false
	}
	return true
}

func FillSegmentGaps[T Segment](in []byte, list *ll.LinkedList[T]) []byte {
	var prev int
	out := make([]byte, 0, len(in))
	orderedSegments := SortAndFilterSegments(list, true, removeBadSegments)
	for i, cur := range orderedSegments {
		if prev > cur.Pos {
			const mask = "decompression should be linear: pos %d and %d collided"
			msg := fmt.Sprintf(mask, orderedSegments[i-1].Pos, cur.Pos)
			panic(msg)
		}
		out = append(out, in[prev:cur.Pos]...)
		originalSize := cur.Segment.GetOriginalSize() / len(cur.Segment.GetPos())
		prev = cur.Pos + originalSize
	}
	out = append(out, in[prev:]...)
	return out
}
