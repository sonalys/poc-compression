package gompressor

func CreateSameCharSegments(in []byte) []byte {
	lenIn := uint32(len(in))
	out := make([]byte, 0, lenIn)
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
			out = append(out, in[prev:index]...)
		}
		list.AppendValue(NewSegment(TypeRepeatSameChar, index, repeatCount, []byte{in[index]}))
		index += uint32(repeatCount) - 1
		prev = index + 1
	}
	if list.Head == nil {
		out = in
	} else if lenIn-prev > 0 {
		out = append(out, in[prev:]...)
	}
	return out
}
