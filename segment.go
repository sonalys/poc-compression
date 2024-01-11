package gompressor

import (
	"bytes"
	"fmt"
	"math"
)

type Segment struct {
	Type           SegmentType
	Repeat         uint16
	Buffer         []byte
	Previous, Next *Segment
	Pos            []uint32
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

// GetCompressionGains returns how many bytes this section is compressing.
func (s *Segment) GetCompressionGains() int64 {
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
func NewSegment(t SegmentType, pos uint32, repeat uint16, buffer []byte) *Segment {
	// flags := meta(0)
	// flags = flags.setIsRepeat2Bytes(repeat > math.MaxUint8)
	// flags = flags.setPosLen(1)
	resp := &Segment{
		Type:   t,
		Repeat: repeat,
		Buffer: buffer,
		Pos:    []uint32{pos},
	}
	return resp
}

// AddPos will append all positions from pos into the current segment,
// it will return error if it overflows the maximum capacity of the segment.
func (s *Segment) AddPos(pos []uint32) (*Segment, error) {
	newLen := len(s.Pos) + len(pos)
	if newLen > 0b11111 {
		// TODO: add better handling for repeating groups.
		return s, fmt.Errorf("len(pos) overflow")
	}
	s.Pos = append(s.Pos, pos...)
	return s, nil
}

// Append adds a segment after the current.
func (s *Segment) Append(next *Segment) *Segment {
	// Finds the tail of the next segment chain.
	cur := next
	for {
		if cur.Next == nil {
			break
		}
		cur = cur.Next
	}
	// Merges the two segment chains.
	cur.Next = s.Next
	s.Next = next
	next.Previous = s
	return next
}

func (s *Segment) Tail() *Segment {
	cur := s
	for {
		if cur.Next == nil {
			break
		}
		cur = cur.Next
	}
	return cur
}
