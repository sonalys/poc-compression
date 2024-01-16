package gompressor

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func Compress(in []byte) *Block {
	size := int64(len(in))
	layers := []func([]byte) *LinkedList[Segment]{
		CreateSameCharSegments,
		CreateRepeatingSegments,
	}
	list := NewLinkedList[Segment]()
	for _, compressionLayer := range layers {
		newSegments := compressionLayer(in)
		Deduplicate(list)
		in = RevertBadSegments(newSegments, size)
		list.Append(newSegments.Head)
	}
	Deduplicate(list)
	return &Block{
		OriginalSize: size,
		List:         list,
		Buffer:       in,
	}
}
