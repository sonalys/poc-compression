package gompressor

import "math"

func CreateSameCharSegments(in []byte) *LinkedList[Segment] {
	lenIn := int64(len(in))
	var prev int64
	list := &LinkedList[Segment]{}
	// finds repetition groups and store them.
	for index := int64(0); index < lenIn; index++ {
		repeatCount := int64(1)
		for j := index + 1; j < lenIn && in[index] == in[j]; j++ {
			repeatCount += 1
			if repeatCount > math.MaxUint16 {
				panic("repeat overflow")
			}
		}
		if repeatCount < 2 {
			continue
		}
		// avoid creating segments with nil buffer.
		if index-prev > 0 {
			list.AppendValue(NewSegment(TypeUncompressed, prev, in[prev:index]))
		}
		list.AppendValue(NewRepeatSegment(index, uint16(repeatCount), []byte{in[index]}))
		index += repeatCount - 1
		prev = index + 1
	}
	// Appends trailing uncompressed segment.
	list.AppendValue(NewSegment(TypeUncompressed, prev, in[prev:]))
	return list
}
