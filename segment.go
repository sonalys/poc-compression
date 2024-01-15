package gompressor

import (
	"bytes"
	"fmt"
	"math"
)

type Segment[S BlockSize] struct {
	Type   SegmentType
	Repeat uint16
	Buffer []byte
	Pos    []S
}

// Decompress returns the segment to it's decompressed state.
func (s *Segment[S]) Decompress() []byte {
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
func (s *Segment[S]) GetOriginalSize() int64 {
	repeat := int64(s.Repeat)
	bufLen := int64(len(s.Buffer))
	posLen := int64(len(s.Pos))
	originalSize := repeat * bufLen * posLen
	return originalSize
}

// GetCompressionGains returns how many bytes this section is compressing.
func (s *Segment[S]) GetCompressionGains() int64 {
	posLen := int64(len(s.Pos))
	bufLen := int64(len(s.Buffer))
	originalSize := s.GetOriginalSize()
	var compressedSize int64 = 5
	// if segment is repeat, then +1 or +2 bytes.
	if s.Type == TypeRepeatSameChar {
		if s.Repeat > math.MaxUint8 {
			compressedSize += 2
		} else {
			compressedSize += 1
		}
	}
	compressedSize += posLen * 4
	compressedSize += bufLen
	gain := originalSize - compressedSize
	return gain
}

// NewSegment creates a new segment.
func NewSegment[S BlockSize](t SegmentType, pos S, repeat uint16, buffer []byte) *Segment[S] {
	resp := &Segment[S]{
		Type:   t,
		Repeat: repeat,
		Buffer: buffer,
		Pos:    []S{pos},
	}
	return resp
}

func (s *Segment[S]) RemovePos(pos S) {
	for i := range s.Pos {
		if s.Pos[i] == pos {
			s.Pos = append(s.Pos[:i], s.Pos[i+1:]...)
			break
		}
	}
}

// AppendPos will append all positions from pos into the current segment,
// it will return error if it overflows the maximum capacity of the segment.
func (s *Segment[S]) AppendPos(pos []S) (*Segment[S], error) {
	newLen := len(s.Pos) + len(pos)
	if newLen > maxSegmentPos {
		return s, fmt.Errorf("len(pos) overflow")
	}
	s.Pos = append(s.Pos, pos...)
	return s, nil
}

func (s *Segment[S]) CanMerge(other *Segment[S]) bool {
	return !(s.Type != other.Type || s.Repeat != other.Repeat || !bytes.Equal(s.Decompress(), other.Decompress()))
}
