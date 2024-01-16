package gompressor

import (
	"bytes"
	"math"
)

type Segment struct {
	Type   SegmentType
	Repeat uint16
	Buffer []byte
	MaxPos int64
	Pos    []int64
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

// GetOriginalSize returns decompressed size for the segment.
func (s *Segment) GetOriginalSize() int64 {
	repeat := int64(s.Repeat)
	bufLen := int64(len(s.Buffer))
	posLen := int64(len(s.Pos))
	originalSize := repeat * bufLen * posLen
	return originalSize
}

func getInt64SegmentBitSize(n int64) int64 {
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

func (s *Segment) GetCompressedSize() int64 {
	posLen := int64(len(s.Pos))
	bufLen := int64(len(s.Buffer))
	var compressedSize int64 = 1
	compressedSize += getInt64SegmentBitSize(bufLen)
	compressedSize += getInt64SegmentBitSize(posLen)
	if s.Type == TypeRepeatSameChar {
		if s.Repeat > math.MaxUint8 {
			compressedSize += 2
		} else {
			compressedSize += 1
		}
	}
	maxPos := s.MaxPos
	compressedSize += posLen * getInt64SegmentBitSize(maxPos)
	compressedSize += bufLen
	return compressedSize
}

// GetCompressionGains returns how many bytes this section is compressing.
func (s *Segment) GetCompressionGains() int64 {
	return s.GetOriginalSize() - s.GetCompressedSize()
}

// NewSegment creates a new segment.
func NewSegment(t SegmentType, pos int64, repeat uint16, buffer []byte) *Segment {
	resp := &Segment{
		Type:   t,
		Repeat: repeat,
		Buffer: buffer,
		MaxPos: pos,
		Pos:    []int64{pos},
	}
	return resp
}

func (s *Segment) RemovePos(pos int64) {
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
func (s *Segment) AppendPos(pos []int64) *Segment {
	for i := range pos {
		if pos[i] > s.MaxPos {
			s.MaxPos = pos[i]
		}
	}
	s.Pos = append(s.Pos, pos...)
	return s
}

func (s *Segment) CanMerge(other *Segment) bool {
	return !(s.Type != other.Type || s.Repeat != other.Repeat || !bytes.Equal(s.Decompress(), other.Decompress()))
}
