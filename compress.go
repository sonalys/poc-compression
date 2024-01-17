package gompressor

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func FillSegmentGaps(buf []byte, list *LinkedList[Segment]) []byte {
	var prev int64
	out := make([]byte, 0, len(buf))
	orderedSegments := sortAndFilterSegments(list, true, func(le *ListEntry[Segment]) bool {
		if le.Value.GetCompressionGains() <= 0 {
			le.Remove()
			return false
		}
		return true
	})
	for _, cur := range orderedSegments {
		if prev > cur.Pos {
			panic("decompression should be linear")
		}
		out = append(out, buf[prev:cur.Pos]...)
		prev = cur.Pos + int64(len(cur.Buffer))*int64(cur.Repeat)
	}
	out = append(out, buf[prev:]...)
	return out
}

func Compress(buf []byte) *Block {
	size := int64(len(buf))
	layers := []func([]byte) (*LinkedList[Segment], []byte){
		CreateSameCharSegments,
		CreateRepeatingSegments2,
	}
	list := NewLinkedList[Segment]()
	for _, compressionLayer := range layers {
		newSegments, out := compressionLayer(buf)
		buf = out
		// note that we are changing buf through each layer.
		// that means different coordinates for each layer.
		list.Append(newSegments.Head)
	}
	// Don't run any optimizations outside, because the coordinates of each layer are relative.
	// So if you merge 2 segments from different layers, in reality they have different coordinates.
	return &Block{
		OriginalSize: size,
		List:         list,
		Buffer:       buf,
	}
}
