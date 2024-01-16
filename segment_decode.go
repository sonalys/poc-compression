package gompressor

var decodeFunc = []func(b []byte, pos int64) (int64, int64){
	func(b []byte, pos int64) (int64, int64) { return int64(b[pos]), pos + 1 },
	func(b []byte, pos int64) (int64, int64) { return int64(decoder.Uint16(b[pos:])), pos + 2 },
	func(b []byte, pos int64) (int64, int64) { return int64(decoder.Uint32(b[pos:])), pos + 4 },
	func(b []byte, pos int64) (int64, int64) { return int64(decoder.Uint64(b[pos:])), pos + 8 },
}

func DecodeSegment(b []byte) (*Segment, int64) {
	var pos int64
	flag, pos := Metadata(b[pos]), pos+1
	cur := Segment{
		Type:   SegmentType(flag.Check(SegmentTypeMask)),
		Repeat: 1,
	}
	if cur.Type == TypeRepeatSameChar {
		switch flag.GetRepSize() {
		case 0:
			cur.Repeat, pos = uint16(b[pos]), pos+1
		case 1:
			cur.Repeat, pos = decoder.Uint16(b[pos:]), pos+2
		}
	}
	var posLen int16
	switch flag.GetPosLenSize() {
	case 0:
		posLen, pos = int16(b[pos]), pos+1
	case 1:
		posLen, pos = int16(decoder.Uint16(b[pos:])), pos+2
	}
	cur.Pos = make([]int64, posLen)
	posDecoder := decodeFunc[flag.GetPosSize()]
	for i := range cur.Pos {
		cur.Pos[i], pos = posDecoder(b, pos)
		if cur.Pos[i] > cur.MaxPos {
			cur.MaxPos = cur.Pos[i]
		}
	}
	var bufLen int64
	switch flag.GetBufLenSize() {
	case 0:
		bufLen, pos = int64(b[pos]), pos+1
	case 1:
		bufLen, pos = int64(decoder.Uint16(b[pos:])), pos+2
	case 2:
		bufLen, pos = int64(decoder.Uint32(b[pos:])), pos+4
	case 3:
		bufLen, pos = int64(decoder.Uint64(b[pos:])), pos+8
	}
	cur.Buffer = b[pos : pos+bufLen]
	return &cur, pos + bufLen
}
