package gompressor

import (
	"bytes"
	"fmt"
	"math"
)

type (
	// My cats are still trying to reach to me about how to increase segments efficiency in disk, so stay tuned.

	Block struct {
		// Size of original buffer.
		Size uint32
		// Head of the dynamic list for buffer segments.
		Head *Segment
		// Buffer of all uncompressed data.
		Buffer []byte
	}

	// Segment
	// memory serialization requires magic to avoid storing as much bytes as possible,
	// while still being able to recover the struct and reconstruct the original buffer.
	Segment struct {
		// segment metadata.
		Metadata meta
		// Repeat indicates how many times buffer is repeated. that's not the same as pos.
		// we are only storing this field in disk if the segment type is typeRepeat.
		Repeat uint16
		// Buffer can hold at maximum 4294967295 bytes, or 4.294967 gigabytes.
		Buffer []byte

		// No need to serialize the fields below on disk.
		// Previous, Next elements in linked list.
		Previous, Next *Segment
		// positions in which this segment repeats itself, max 32 positions.
		Pos []uint32 // todo: no need to hold pos if there is only 1 value.
	}
)

func (s *Segment) Decompress() []byte {
	switch s.Metadata.getType() {
	case typeUncompressed:
		return s.Buffer
	case typeRepeat:
		return bytes.Repeat(s.Buffer, int(s.Repeat))
	default:
		panic("invalid segment type")
	}
}

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
	if s.Metadata.getType() == typeRepeat {
		if s.Metadata.isRepeat2Bytes() {
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
func NewSegment(t segType, pos uint32, repeat uint16, buffer []byte) *Segment {
	flags := meta(0)
	flags = flags.setIsRepeat2Bytes(repeat > math.MaxUint8)
	flags = flags.setPosLen(1)
	flags = flags.setType(t)
	resp := &Segment{
		Metadata: flags,
		Repeat:   repeat,
		Buffer:   buffer,
		Pos:      []uint32{pos},
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
	s.Metadata = s.Metadata.setPosLen(uint8(newLen))
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
			if !bytes.Equal(cur.Buffer, iter.Buffer) || cur.Repeat != iter.Repeat || cur.Metadata.getType() != iter.Metadata.getType() {
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
