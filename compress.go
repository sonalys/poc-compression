package gompressor

import "math"

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func Compress(in []byte) *Block {
	if len(in) > math.MaxUint32 {
		panic("input is over 4294967295 bytes long")
	}
	size := uint32(len(in))

	layers := []func([]byte) *LinkedList[Segment]{
		CreateSameCharSegments,
		CreateRepeatingSegments,
	}

	list := NewLinkedList[Segment]().AppendValue(nil)
	for _, layer := range layers {
		layer := layer(in)
		in = RevertBadSegments(layer, size)
		list.Tail.Append(layer.Head)
	}
	list.Head.Remove()
	Deduplicate(list)
	return &Block{
		Size:   size,
		List:   list,
		Buffer: in,
	}
}
