package gompressor

import "fmt"

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func FillSegmentGaps(buf []byte, list *LinkedList[*Segment]) []byte {
	var prev int
	out := make([]byte, 0, len(buf))
	orderedSegments := sortAndFilterSegments(list, true, func(le *ListEntry[*Segment]) bool {
		if le.Value.GetCompressionGains() <= 0 {
			le.Remove()
			return false
		}
		return true
	})
	for i, cur := range orderedSegments {
		if prev > cur.Pos {
			const mask = "decompression should be linear: pos %d and %d collided with size %d"
			msg := fmt.Sprintf(mask, orderedSegments[i-1].Pos, cur.Pos, cur.ByteCount)
			panic(msg)
		}
		out = append(out, buf[prev:cur.Pos]...)
		prev = cur.Pos + cur.ByteCount*int(cur.Repeat)
	}
	out = append(out, buf[prev:]...)
	return out
}

func Compress(buf []byte) *Block {
	size := len(buf)
	layers := []func([]byte) (*LinkedList[*Segment], []byte){
		CreateSameCharSegments,
		// CreateRepeatingSegments,
	}
	list := NewLinkedList[*Segment]()
	for _, compressionLayer := range layers {
		var newSegments *LinkedList[*Segment]
		// note that we are changing buf through each layer.
		// that means different coordinates for each layer.
		newSegments, buf = compressionLayer(buf)
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
