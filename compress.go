package gompressor

import "math"

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func Compress(in []byte) *Block {
	if len(in) > math.MaxUint32 {
		panic("input is over 4294967295 bytes long")
	}
	size := uint32(len(in))
	// Layer 1 - Remove same char repeat.
	layer1 := CreateSameCharSegments(in)
	out := RevertBadSegments(layer1, size)
	// Layer 2 - Remove group repeat.
	layer2 := CreateRepeatingSegments(out)
	out = RevertBadSegments(layer2, size)
	// Build final segment list.
	layer2.Tail.Append(layer1.Head)
	Deduplicate(layer2)
	return &Block{
		Size:   size,
		List:   layer2,
		Buffer: out,
	}
}
