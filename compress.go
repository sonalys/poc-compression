package gompressor

import "math"

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func Compress(in []byte) *Block {
	if len(in) > math.MaxUint32 {
		panic("input is over 4294967295 bytes long")
	}
	//out, tail := CreateSameCharSegments(in)
	out, head := CreateRepeatingSegments(in)
	// head.Tail().Append(tail)
	b := Block{
		Size:   uint32(len(in)),
		Head:   head,
		Buffer: out,
	}
	b.Optimize()
	return &b
}
