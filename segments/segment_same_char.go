package segments

import (
	"bytes"
	"math"
)

type SegmentSameChar struct {
	char   byte
	repeat int
	pos    []int
	maxPos int
}

func NewRepeatSegment(repeat int, char byte, pos ...int) *SegmentSameChar {
	resp := &SegmentSameChar{
		char:   char,
		repeat: repeat,
	}
	return resp.appendPos(pos...)
}

func (s *SegmentSameChar) appendPos(pos ...int) *SegmentSameChar {
	for i := range pos {
		if pos[i] > s.maxPos {
			s.maxPos = pos[i]
		}
	}
	s.pos = append(s.pos, pos...)
	return s
}

func (s *SegmentSameChar) Decompress() []byte {
	return bytes.Repeat([]byte{s.char}, s.repeat)
}

func getStorageByteSize(n int) int {
	switch {
	case n > math.MaxUint32:
		return 8
	case n > math.MaxUint16:
		return 4
	case n > math.MaxUint8:
		return 2
	default:
		return 1
	}
}

func calculateSameCharCompressedSize(posLen, repeat, maxPos int) int {
	var compressedSize int = 1
	if posLen > 1 {
		compressedSize += getStorageByteSize(posLen)
	}
	if repeat > math.MaxUint8 {
		compressedSize += 2
	} else {
		compressedSize += 1
	}
	compressedSize += posLen * getStorageByteSize(maxPos)
	compressedSize += 1
	return compressedSize
}

func (s *SegmentSameChar) getCompressedSize() int {
	return calculateSameCharCompressedSize(len(s.pos), s.repeat, s.maxPos)
}

func (s *SegmentSameChar) GetOriginalSize() int {
	posLen := len(s.pos)
	originalSize := s.repeat * posLen
	return originalSize
}

func (s *SegmentSameChar) GetCompressionGains() int {
	return s.GetOriginalSize() - s.getCompressedSize()
}

func (s *SegmentSameChar) GetPos() []int {
	return s.pos
}

func (s *SegmentSameChar) GetType() SegmentType {
	return TypeSameChar
}

func (s *SegmentSameChar) Encode() []byte {
	buffer := make([]byte, 0, s.getCompressedSize())
	posLen := len(s.pos)

	meta := MetaSameChar{
		Type:       TypeSameChar,
		SinglePos:  posLen == 1,
		RepeatSize: NewMaxSize(s.repeat),
		PosLenSize: NewMaxSize(posLen),
		PosSize:    NewMaxSize(s.maxPos),
	}

	if meta.RepeatSize >= 2 {
		panic("fuck")
	}

	buffer = append(buffer, meta.ToByte())
	buffer = encodingFunc[meta.RepeatSize](buffer, s.repeat)
	if meta.SinglePos {
		buffer = encodingFunc[meta.PosSize](buffer, s.pos[0])
	} else {
		buffer = encodingFunc[meta.PosLenSize](buffer, posLen)
		buffer = append(buffer, encodePos(s.maxPos, s.pos)...)
	}
	buffer = append(buffer, s.char)
	return buffer
}

func DecodeSameChar(b []byte) (*SegmentSameChar, int) {
	var pos int
	meta, pos := NewSameCharMeta(b[pos]), pos+1
	cur := SegmentSameChar{}
	cur.repeat, pos = decodeFunc[meta.RepeatSize](b, pos)
	if meta.SinglePos {
		var segPos int
		segPos, pos = decodeFunc[meta.PosSize](b, pos)
		cur.appendPos(segPos)
	} else {
		cur.maxPos, pos, cur.pos = decodePos(b, pos, meta.PosLenSize, meta.PosSize)
	}
	cur.char, pos = b[pos], pos+1
	return &cur, pos
}
