package segments

type Segment interface {
	Decompress() []byte
	Encode() []byte
	GetCompressionGains() int
	GetOriginalSize() int
	GetPos() []int
	GetType() SegmentType
}

func DecodeSegment(b []byte) (Segment, int) {
	t := getSegmentType(b[0])
	switch t {
	case TypeSameChar:
		return DecodeSameChar(b)
	case TypeGroup:
		return DecodeGroup(b)
	default:
		panic("unknown segment type")
	}
}
