package gompressor

import "sort"

type SegmentPosMap struct {
	Pos uint32
	*Segment
}

func sortAndFilterSegments(head *Segment, filters ...func(*Segment) bool) []SegmentPosMap {
	out := make([]SegmentPosMap, 0, 500)
	head.ForEach(func(s *Segment) {
		for _, filter := range filters {
			if !filter(s) {
				return
			}
		}
		for _, pos := range s.Pos {
			out = append(out, SegmentPosMap{
				Pos:     pos,
				Segment: s,
			})
		}
	})
	sort.Slice(out, func(i, j int) bool {
		// We layer the logic by segment type, so some segments should decompress first than others.
		// if out[i].Type != out[j].Type {
		// 	return out[i].Type < out[j].Type
		// }
		return out[i].Pos < out[j].Pos
	})
	return out
}

func Decompress(b *Block) []byte {
	out := make([]byte, b.Size)
	copy(out, b.Buffer)
	for _, cur := range sortAndFilterSegments(b.Head) {
		buf := cur.Decompress()
		lenBuf := uint32(len(buf))
		// right-shift data.
		copy(out[cur.Pos+lenBuf:], out[cur.Pos:])
		// copy out decompressed buf into out[pos].
		copy(out[cur.Pos:], buf)
	}
	return out
}
