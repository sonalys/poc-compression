package gompressor

import "math"

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func Compress(in []byte) *Block {
	if len(in) > math.MaxUint32 {
		panic("input is over 4294967295 bytes long")
	}
	size := uint32(len(in))
	//out, tail := CreateSameCharSegments(in)
	// head.Tail().Append(tail)
	head := CreateRepeatingSegments(in)
	head = head.Deduplicate()
	head, uncompressedBuffer := head.RevertBadSegments(size)
	return &Block{
		Size:   size,
		Head:   head,
		Buffer: uncompressedBuffer,
	}
}
