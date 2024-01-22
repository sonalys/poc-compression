package gompressor

var decodeFunc = []func(b []byte, pos int) (int, int){
	func(b []byte, pos int) (int, int) { return int(b[pos]), pos + 1 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint16(b[pos:])), pos + 2 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint32(b[pos:])), pos + 4 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint64(b[pos:])), pos + 8 },
}

func decodePos(b []byte, pos int, posLenSize, posSize MaxSize) (int, int, []int) {
	var posLen int16
	switch posLenSize {
	case MaxSizeUint8:
		posLen, pos = int16(b[pos]), pos+1
	case MaxSizeUint16:
		posLen, pos = int16(decoder.Uint16(b[pos:])), pos+2
	default:
		panic("invalid posSize")
	}
	var maxPos int
	posList := make([]int, posLen)
	posDecoder := decodeFunc[posSize]
	for i := range posList {
		posList[i], pos = posDecoder(b, pos)
		if posList[i] > maxPos {
			maxPos = posList[i]
		}
	}
	return maxPos, pos, posList
}

func decodeSameChar(b []byte) (*Segment, int) {
	var pos int
	meta, pos := NewMeta2(b[pos]), pos+1
	cur := Segment{
		Type:      TypeRepeatSameChar,
		ByteCount: 1,
		BitMask:   0xff,
	}
	cur.Repeat, pos = decodeFunc[meta.RepeatSize](b, pos)
	cur.MaxPos, pos, cur.Pos = decodePos(b, pos, meta.PosLenSize, meta.PosSize)
	cur.Buffer, pos = b[pos:pos+1], pos+1
	return &cur, pos
}

func decodeRepeatingGroup(b []byte) (*Segment, int) {
	var pos int
	meta, pos := NewMeta2(b[pos]), pos+1
	cur := Segment{
		Type:       TypeRepeatingGroup,
		Repeat:     1,
		BitMask:    0xff,
		InvertMask: meta.InvertBitmask,
	}
	cur.BitMask, pos = b[pos], pos+1
	cur.MaxPos, pos, cur.Pos = decodePos(b, pos, meta.PosLenSize, meta.PosSize)
	cur.ByteCount, pos = decodeFunc[meta.BufLenSize](b, pos)
	maskSize := Count1Bits(cur.BitMask)
	if maskSize == 8 || maskSize == 0 {
		cur.Buffer = b[pos : pos+cur.ByteCount]
		return &cur, pos + cur.ByteCount
	}
	compLen := (maskSize*cur.ByteCount + 8 - 1) / 8
	cur.Buffer, pos = b[pos:pos+compLen], pos+compLen
	return &cur, pos
}

func DecodeSegment(b []byte) (*Segment, int) {
	t := SegmentType(b[0] & 0b1)
	switch t {
	case TypeRepeatSameChar:
		return decodeSameChar(b)
	case TypeRepeatingGroup:
		return decodeRepeatingGroup(b)
	default:
		panic("unknown segment type")
	}
}
