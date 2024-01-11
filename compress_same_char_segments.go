package gompressor

func CreateSameCharSegments(in []byte) *LinkedList[Segment] {
	lenIn := uint32(len(in))
	var prev uint32
	list := &LinkedList[Segment]{}
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
			list.AppendValue(NewSegment(TypeUncompressed, prev, 1, in[prev:index]))
		}
		list.AppendValue(NewSegment(TypeRepeatSameChar, index, repeatCount, []byte{in[index]}))
		index += uint32(repeatCount) - 1
		prev = index + 1
	}
	if list.Head == nil {
		list.AppendValue(NewSegment(TypeUncompressed, 0, 1, in))
	} else if lenIn-prev > 0 {
		list.AppendValue(NewSegment(TypeUncompressed, prev, 1, in[prev:]))
	}
	return list
}
