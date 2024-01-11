package gompressor

import "sort"

type SegmentPosMap struct {
	Pos uint32
	*Segment
}

func sortAndFilterSegments(list *LinkedList[Segment], sortType bool, filters ...func(*ListEntry[Segment]) bool) []SegmentPosMap {
	out := make([]SegmentPosMap, 0, list.Len)
	cur := list.Head
	for {
		if cur == nil {
			break
		}
		curValue := cur.Value
		for _, filter := range filters {
			if !filter(cur) {
				goto final
			}
		}
		for _, pos := range curValue.Pos {
			out = append(out, SegmentPosMap{
				Pos:     pos,
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

func Decompress(b *Block) []byte {
	out := make([]byte, b.Size)
	copy(out, b.Buffer)
	for _, cur := range sortAndFilterSegments(b.List, true) {
		buf := cur.Decompress()
		lenBuf := uint32(len(buf))
		// right-shift data.
		copy(out[cur.Pos+lenBuf:], out[cur.Pos:])
		// copy out decompressed buf into out[pos].
		copy(out[cur.Pos:], buf)
	}
	return out
}
