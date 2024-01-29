package segments

import "github.com/sonalys/gompressor/bitbuffer"

type Segment interface {
	Decompress(pos int) []byte
	Encode(w *bitbuffer.BitBuffer)
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
	case TypeMasked:
		return DecodeMasked(b)
	default:
		panic("unknown segment type")
	}
}
