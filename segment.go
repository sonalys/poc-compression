package gompressor

import (
	"bytes"
	"fmt"
	"math"
)

type (
	Block struct {
		Size   uint32
		Head   *Segment
		Buffer []byte
	}

	Segment struct {
		Type           SegmentType
		Repeat         uint16
		Buffer         []byte
		Previous, Next *Segment
		Pos            []uint32
	}
)

// Decompress returns the segment to it's decompressed state.
func (s *Segment) Decompress() []byte {
	switch s.Type {
	case TypeUncompressed:
		return s.Buffer
	case TypeRepeat:
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
	if s.Type == TypeRepeat {
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

// Add append a new segment to the linked list.
func (s *Segment) Add(next *Segment) *Segment {
	s.Next = next
	next.Previous = s
	return next
}

// Remove dereferences this segment from the linked list.
func (b *Block) Remove(s *Segment) {
	if s.Previous == nil {
		b.Head = s.Next
	} else {
		s.Previous.Next = s.Next
	}
	if s.Next != nil {
		s.Next.Previous = s.Previous
	}
}

// Deduplicate will find segments that are identical, besides position, and merge them.
func (b *Block) Deduplicate() {
	cur := b.Head
	for {
		if cur == nil {
			break
		}
		iter := cur
		for {
			if iter = iter.Next; iter == nil {
				break
			}
			if !bytes.Equal(cur.Buffer, iter.Buffer) || cur.Repeat != iter.Repeat || cur.Type != iter.Type {
				continue
			}
			// if pos doesn't overflow, we continue with the merge operation.
			if _, err := cur.AddPos(iter.Pos); err == nil {
				b.Remove(iter)
			}
		}
		cur = cur.Next
	}
}

// Optimize is responsible for finding segments that are causing byte compression gain to be negative, and try to
// revert it.
func (b *Block) Optimize() {
	cur := b.Head
	// For making the logic easier on the POC, we just use this slice to sort by position.
	segMap := make([]*SegmentPosMap, b.Size)
	for {
		if cur == nil {
			break
		}
		// If we are not gaining any delta size, we just move it to the uncompressed buffer.
		if cur.GetCompressionGains() <= 0 {
			for _, pos := range cur.Pos {
				segMap[pos] = &SegmentPosMap{
					Pos:     pos,
					Segment: cur,
				}
			}
			b.Remove(cur)
		}
		cur = cur.Next
	}

	for _, entry := range segMap {
		if entry == nil {
			continue
		}
		cur, pos := entry.Segment, entry.Pos
		segBuf := cur.Decompress()
		bufLen := uint32(len(b.Buffer))
		if pos < bufLen {
			panic("buffer optimization should be linear")
		}
		b.Buffer = append(b.Buffer, segBuf...)
	}
}

// FindSegment finds the segment that contains pos.
func (s *Segment) FindSegment(pos uint32) (*Segment, bool) {
	cur := s
	for {
		for _, curPos := range cur.Pos {
			if curPos <= pos && curPos+uint32(cur.Repeat)*uint32(len(cur.Buffer)) > pos {
				return cur, true
			}
		}
		if cur.Next == nil {
			break
		}
		cur = cur.Next
	}
	return nil, false
}

func (b *Block) ForEach(f func(*Segment)) {
	cur := b.Head
	for {
		if cur == nil {
			break
		}
		f(cur)
		cur = cur.Next
	}
}
