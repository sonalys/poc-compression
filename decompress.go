package gompressor

type SegmentPosMap struct {
	Pos uint32
	*Segment
}

func sortSegmentsByPos(b *Block) []*SegmentPosMap {
	segMap := make([]*SegmentPosMap, b.Size)
	b.ForEach(func(s *Segment) {
		for _, pos := range s.Pos {
			segMap[pos] = &SegmentPosMap{
				Pos:     pos,
				Segment: s,
			}
		}
	})
	sortedSegments := make([]*SegmentPosMap, 0, b.Size)
	for i := range segMap {
		if segMap[i] == nil {
			continue
		}
		sortedSegments = append(sortedSegments, segMap[i])
	}
	return sortedSegments
}

func Decompress(b *Block) []byte {
	out := make([]byte, b.Size)
	segments := sortSegmentsByPos(b)
	if len(segments) == 0 {
		return b.Buffer
	}
	copy(out, b.Buffer)
	for _, cur := range segments {
		buf := cur.Decompress()
		lenBuf := uint32(len(buf))
		// right-shift data.
		copy(out[cur.Pos+lenBuf:], out[cur.Pos:])
		// copy out decompressed buf into out[pos].
		copy(out[cur.Pos:], buf)
	}
	return out
}
