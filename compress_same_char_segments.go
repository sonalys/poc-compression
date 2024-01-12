package gompressor

func CreateSameCharSegments[S BlockSize](in []byte) *LinkedList[Segment[S]] {
	lenIn := S(len(in))
	var prev S
	list := &LinkedList[Segment[S]]{}
	// finds repetition groups and store them.
	for index := S(0); index < lenIn; index++ {
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
		index += S(repeatCount) - 1
		prev = index + 1
	}
	if list.Head == nil {
		list.AppendValue(NewSegment[S](TypeUncompressed, 0, 1, in))
	} else if lenIn-prev > 0 {
		list.AppendValue(NewSegment(TypeUncompressed, prev, 1, in[prev:]))
	}
	return list
}
