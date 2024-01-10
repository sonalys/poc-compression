package gompressor

import "math"

// TODO: Test bloom-filters for regenerating a byte dictionary
// try to create 2 or 3 filters, multiplying by prime numbers to get more precision.

func Compress(in []byte) *Block {
	if len(in) > math.MaxUint32 {
		panic("input is over 4294967295 bytes long")
	}
	lenIn := uint32(len(in))
	b := Block{
		Size: lenIn,
		Head: &Segment{},
	}
	// prev is a cursor to the last position before a repeating group
	var prev uint32
	// cur is a cursor to the head of the block's segments.
	cur := b.Head
	// finds repetition groups and store them.
	for index := uint32(0); index < lenIn; index++ {
		repeatCount := uint16(1)
		for j := index + 1; j < lenIn && in[index] == in[j]; j++ {
			repeatCount += 1
			if repeatCount == 0 {
				panic("repeat overflow")
			}
		}
		if repeatCount < 2 {
			continue
		}
		// avoid creating segments with nil buffer.
		if index-prev > 0 {
			cur = cur.Add(NewSegment(typeUncompressed, prev, 1, in[prev:index]))
		}
		cur = cur.Add(NewSegment(typeRepeat, index, repeatCount, []byte{in[index]}))
		index += uint32(repeatCount) - 1
		prev = index + 1
	}
	// Fix head.
	b.Remove(b.Head)
	if b.Head == nil {
		b.Head = NewSegment(typeUncompressed, 0, 1, in)
	} else if lenIn-prev > 0 {
		cur.Add(NewSegment(typeUncompressed, prev, 1, in[prev:]))
	}

	b.Deduplicate()
	b.Optimize()
	return &b
}
