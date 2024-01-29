package segments

import (
	"bytes"
	"math"

	"github.com/sonalys/gompressor/bitbuffer"
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

func (s *SegmentSameChar) Decompress(pos int) []byte {
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
	var compressedSize int
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

func (s *SegmentSameChar) Encode(w *bitbuffer.BitBuffer) {
	w.WriteBuffer(encodingFunc[1](nil, s.repeat))
	for i, pos := range s.pos {
		w.WriteBuffer(encodingFunc[2](nil, pos))
		if i == len(s.pos)-1 {
			w.Write(0b1, 1)
			break
		}
		w.Write(0b0, 1)
	}
	w.Write(s.char, 8)
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
