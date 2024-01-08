package gompressor

import "math"

func Compress(in []byte) *block {
	if len(in) > math.MaxUint32 {
		panic("input is over 4294967295 bytes long")
	}
	lenIn := uint32(len(in))
	b := block{
		Size: lenIn,
	}
	// prev is a cursor to the last position before a repeating group
	var prev uint32
	head := &Segment{}
	// cur is a cursor to the head of the block's segments.
	cur := head
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
	head = head.Next
	if head == nil {
		head = NewSegment(typeUncompressed, 0, 1, in)
	} else if lenIn-prev > 0 {
		cur.Add(NewSegment(typeUncompressed, prev, 1, in[prev:]))
	}

	head.Deduplicate()
	head.Optimize()

	b.Segments = head.GetOrderedSegments()
	return &b
}
