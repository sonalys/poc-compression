package gompressor

import "sort"

type SegmentPosMap struct {
	Pos int
	*Segment
	Entry *ListEntry[*Segment]
}

func sortAndFilterSegments(list *LinkedList[*Segment], sortType bool, filters ...func(*ListEntry[*Segment]) bool) []SegmentPosMap {
	out := make([]SegmentPosMap, 0, list.Len)
	cur := list.Head
	for {
		if cur == nil {
			break
		}
		curValue := cur.Value
		if len(cur.Value.Pos) == 0 {
			goto final
		}
		for _, filter := range filters {
			if !filter(cur) {
				goto final
			}
		}
		for _, pos := range curValue.Pos {
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
			// We layer the logic by segment type, so some segments should decompress first than others.
			if out[i].Type != out[j].Type {
				return out[i].Type < out[j].Type
			}
		}
		return out[i].Pos < out[j].Pos
	})
	return out
}
