package gompressor

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func Compress(buf []byte) *Block {
	size := int64(len(buf))
	layers := []func([]byte) *LinkedList[Segment]{
		CreateSameCharSegments,
		CreateRepeatingSegments,
	}
	list := NewLinkedList[Segment]()
	for _, compressionLayer := range layers {
		newSegments := compressionLayer(buf)
		Deduplicate(newSegments)
		// note that we are changing buf through each layer.
		// that means different coordinates for each layer.
		buf = RevertBadSegments(newSegments, size)
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
