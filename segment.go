package gompressor

import (
	"bytes"
	"math"
)

type Segment struct {
	Type   SegmentType
	Buffer []byte
	Pos    []int
	MaxPos int
	Repeat int
}

// Decompress returns the segment to it's decompressed state.
func (s *Segment) Decompress() []byte {
	switch s.Type {
	case TypeUncompressed, TypeRepeatingGroup:
		return s.Buffer
	case TypeRepeatSameChar:
		return bytes.Repeat(s.Buffer, int(s.Repeat))
	default:
		panic("invalid segment type")
	}
}

func GetOriginalSize(t SegmentType, repeat int, posLen, bufLen int) int {
	switch t {
	case TypeRepeatSameChar:
		return repeat * bufLen * posLen
	default:
		return bufLen * posLen
	}
}

// GetOriginalSize returns decompressed size for the segment.
func (s *Segment) GetOriginalSize() int {
	return GetOriginalSize(s.Type, s.Repeat, len(s.Pos), len(s.Buffer))
}

func getintSegmentBitSize(n int) int {
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

func GetCompressedSize(t SegmentType, repeat, maxPos, posLen, size int) int {
	var compressedSize int = 1
	compressedSize += getintSegmentBitSize(size)
	compressedSize += getintSegmentBitSize(posLen)
	if t == TypeRepeatSameChar {
		if repeat > math.MaxUint8 {
			compressedSize += 2
		} else {
			compressedSize += 1
		}
	}
	compressedSize += posLen * getintSegmentBitSize(maxPos)
	compressedSize += size
	return compressedSize
}

func (s *Segment) GetCompressedSize() int {
	posLen := len(s.Pos)
	bufLen := len(s.Buffer)
	return GetCompressedSize(s.Type, s.Repeat, s.MaxPos, posLen, bufLen)
}

func (s *Segment) VerifySegment(buf []byte) {
	size := len(s.Buffer)
	for _, pos := range s.Pos {
		for offset := 0; offset < size; offset++ {
			if buf[pos+offset] != s.Buffer[offset] {
				panic("segment position is not matching buffer")
			}
		}
	}
}

// GetCompressionGains returns how many bytes this section is compressing.
func (s *Segment) GetCompressionGains() int {
	return s.GetOriginalSize() - s.GetCompressedSize()
}

// NewSegment creates a new segment.
func NewSegment(t SegmentType, buffer []byte, pos ...int) *Segment {
	resp := &Segment{
		Type:   t,
		Buffer: buffer,
		Repeat: 1,
	}
	return resp.AppendPos(pos...)
}

// NewSegment creates a new segment.
func NewRepeatSegment(repeat int, buffer []byte, pos ...int) *Segment {
	resp := &Segment{
		Type:   TypeRepeatSameChar,
		Buffer: buffer,
		Repeat: repeat,
	}
	return resp.AppendPos(pos...)
}

func (s *Segment) RemovePos(pos int) {
	if s.MaxPos == pos {
		s.MaxPos = Max(s.Pos)
	}
	for i := range s.Pos {
		if s.Pos[i] == pos {
			s.Pos = append(s.Pos[:i], s.Pos[i+1:]...)
			break
		}
	}
}

// AppendPos will append all positions from pos into the current segment,
// it will return error if it overflows the maximum capacity of the segment.
func (s *Segment) AppendPos(pos ...int) *Segment {
	for i := range pos {
		if pos[i] > s.MaxPos {
			s.MaxPos = pos[i]
		}
	}
	s.Pos = append(s.Pos, pos...)
	return s
}

func (s *Segment) CanMerge(other *Segment) bool {
	return len(s.Buffer) == len(other.Buffer) && s.Repeat == other.Repeat && bytes.Equal(s.Buffer, other.Buffer)
}
