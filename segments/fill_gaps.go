package segments

import (
	"fmt"
	"sort"

	"github.com/sonalys/gompressor/linkedlist"
)

type SegmentPosMap struct {
	Pos int
	Segment
	Entry *linkedlist.ListEntry[Segment]
}

func SortAndFilterSegments(list *linkedlist.LinkedList[Segment], sortType bool, filters ...func(*linkedlist.ListEntry[Segment]) bool) []SegmentPosMap {
	out := make([]SegmentPosMap, 0, list.Len)
	cur := list.Head
	for {
		if cur == nil {
			break
		}
		curValue := cur.Value
		if len(cur.Value.GetPos()) == 0 {
			goto final
		}
		for _, filter := range filters {
			if !filter(cur) {
				goto final
			}
		}
		for _, pos := range curValue.GetPos() {
			out = append(out, SegmentPosMap{
				Pos:     pos,
				Entry:   cur,
				Segment: curValue,
			})
		}
	final:
		cur = cur.Next
	}
	sort.Slice(out, func(i, j int) bool {
		if sortType {
			t1, t2 := out[i].GetType(), out[j].GetType()
			// We layer the logic by segment type, so some segments should decompress first than others.
			if t1 != t2 {
				return t1 < t2
			}
		}
		return out[i].Pos < out[j].Pos
	})
	return out
}

func FillSegmentGaps(buf []byte, list *linkedlist.LinkedList[Segment]) []byte {
	var prev int
	out := make([]byte, 0, len(buf))
	orderedSegments := SortAndFilterSegments(list, true, func(le *linkedlist.ListEntry[Segment]) bool {
		if le.Value.GetCompressionGains() <= 0 {
			le.Remove()
			return false
		}
		return true
	})
	for i, cur := range orderedSegments {
		if prev > cur.Pos {
			const mask = "decompression should be linear: pos %d and %d collided"
			msg := fmt.Sprintf(mask, orderedSegments[i-1].Pos, cur.Pos)
			panic(msg)
		}
		out = append(out, buf[prev:cur.Pos]...)
		prev = cur.Pos + cur.Segment.GetOriginalSize()/len(cur.Segment.GetPos())
	}
	out = append(out, buf[prev:]...)
	return out
}
