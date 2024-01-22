package gompressor

var encodingFunc = []func(buffer []byte, value int) []byte{
	func(buffer []byte, value int) []byte { return append(buffer, byte(value)) },
	func(buffer []byte, value int) []byte { return encoder.AppendUint16(buffer, uint16(value)) },
	func(buffer []byte, value int) []byte { return encoder.AppendUint32(buffer, uint32(value)) },
	func(buffer []byte, value int) []byte { return encoder.AppendUint64(buffer, uint64(value)) },
}

func encodePos(maxPos int, posList []int) []byte {
	posSize := NewMaxSize(maxPos)
	buffer := make([]byte, 0, int(posSize+1)*len(posList))
	for i := range posList {
		buffer = encodingFunc[posSize](buffer, posList[i])
	}
	return buffer
}

func encodeSameChar(s *Segment) []byte {
	buffer := make([]byte, 0, s.GetCompressedSize())
	posLen := len(s.Pos)

	meta := Meta{
		Type:       s.Type,
		RepeatSize: NewMaxSize(s.Repeat),
		PosLenSize: NewMaxSize(posLen),
		PosSize:    NewMaxSize(s.MaxPos),
	}

	buffer = append(buffer, meta.ToByte())
	buffer = encodingFunc[meta.RepeatSize](buffer, s.Repeat)
	buffer = encodingFunc[meta.PosLenSize](buffer, posLen)
	buffer = append(buffer, encodePos(s.MaxPos, s.Pos)...)
	buffer = append(buffer, s.Buffer[0])
	return buffer
}

func encodeRepeatingGroup(s *Segment) []byte {
	buffer := make([]byte, 0, s.GetCompressedSize())
	posLen := len(s.Pos)

	meta := Meta{
		Type:          s.Type,
		InvertBitmask: s.InvertMask,
		PosLenSize:    NewMaxSize(posLen),
		PosSize:       NewMaxSize(s.MaxPos),
		BufLenSize:    NewMaxSize(s.ByteCount),
	}

	buffer = append(buffer, meta.ToByte())
	buffer = append(buffer, s.BitMask)
	buffer = encodingFunc[meta.PosLenSize](buffer, posLen)
	buffer = append(buffer, encodePos(s.MaxPos, s.Pos)...)
	buffer = encodingFunc[meta.BufLenSize](buffer, s.ByteCount)
	buffer = append(buffer, s.Buffer...)
	return buffer
}

func (s *Segment) Encode() []byte {
	switch s.Type {
	case TypeRepeatSameChar:
		return encodeSameChar(s)
	case TypeRepeatingGroup:
		return encodeRepeatingGroup(s)
	default:
		panic("unknown segment type")
	}
}
