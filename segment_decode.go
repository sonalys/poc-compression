package gompressor

var decodeFunc = []func(b []byte, pos int) (int, int){
	func(b []byte, pos int) (int, int) { return int(b[pos]), pos + 1 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint16(b[pos:])), pos + 2 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint32(b[pos:])), pos + 4 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint64(b[pos:])), pos + 8 },
}

func DecodeSegment(b []byte) (*Segment, int) {
	var pos int
	flag, pos := Metadata(b[pos]), pos+1
	cur := Segment{
		Type:       SegmentType(flag.Check(SegmentTypeMask)),
		Repeat:     1,
		BitMask:    0xff,
		InvertMask: flag.GetInvertBitMask() != 0,
	}
	if cur.Type == TypeRepeatSameChar {
		switch flag.GetRepSize() {
		case 0:
			cur.Repeat, pos = int(b[pos]), pos+1
		case 1:
			cur.Repeat, pos = int(decoder.Uint16(b[pos:])), pos+2
		}
	} else {
		cur.BitMask, pos = b[pos], pos+1
	}
	var posLen int16
	switch flag.GetPosLenSize() {
	case 0:
		posLen, pos = int16(b[pos]), pos+1
	case 1:
		posLen, pos = int16(decoder.Uint16(b[pos:])), pos+2
	}
	cur.Pos = make([]int, posLen)
	posDecoder := decodeFunc[flag.GetPosSize()]
	for i := range cur.Pos {
		cur.Pos[i], pos = posDecoder(b, pos)
		if cur.Pos[i] > cur.MaxPos {
			cur.MaxPos = cur.Pos[i]
		}
	}
	switch flag.GetBufLenSize() {
	case 0:
		cur.ByteCount, pos = int(b[pos]), pos+1
	case 1:
		cur.ByteCount, pos = int(decoder.Uint16(b[pos:])), pos+2
	case 2:
		cur.ByteCount, pos = int(decoder.Uint32(b[pos:])), pos+4
	case 3:
		cur.ByteCount, pos = int(decoder.Uint64(b[pos:])), pos+8
	}
	maskSize := Count1Bits(cur.BitMask)
	if maskSize == 8 || maskSize == 0 {
		cur.Buffer = b[pos : pos+cur.ByteCount]
		return &cur, pos + cur.ByteCount
	}
	compLen := (maskSize*cur.ByteCount + 8 - 1) / 8
	cur.Buffer, pos = b[pos:pos+compLen], pos+compLen
	return &cur, pos
}
