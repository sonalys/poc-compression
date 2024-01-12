package gompressor

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func Compress[S BlockSize](in []byte) *Block[S] {
	size := S(len(in))
	if int64(len(in)) > int64(size) {
		panic("size overflow on compress input")
	}
	layers := []func([]byte) *LinkedList[Segment[S]]{
		CreateSameCharSegments[S],
		CreateRepeatingSegments[S],
	}
	list := NewLinkedList[Segment[S]]().AppendValue(nil)
	for _, layer := range layers {
		layer := layer(in)
		in = RevertBadSegments[S](layer, size)
		list.Tail.Append(layer.Head)
	}
	list.Head.Remove()
	Deduplicate(list)
	return &Block[S]{
		Size:   size,
		List:   list,
		Buffer: in,
	}
}
