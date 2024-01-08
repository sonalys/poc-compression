package gompressor

import (
	"bytes"
	"fmt"
	"math"
	"sort"
)

type (
	// My cats are still trying to reach to me about how to increase segments efficiency in disk, so stay tuned.

	block struct {
		size     uint32
		segments []diskSegment
	}

	// segment
	// memory serialization requires magic to avoid storing as much bytes as possible,
	// while still being able to recover the struct and reconstruct the original buffer.
	segment struct {
		// segment metadata.
		flags meta
		// repeat indicates how many times buffer is repeated. that's not the same as pos.
		// we are only storing this field in disk if the segment type is typeRepeat.
		repeat uint16
		// buffer can hold at maximum 4294967295 bytes, or 4.294967 gigabytes.
		buffer []byte

		// No need to serialize the fields below on disk.
		// previous, next elements in linked list.
		previous, next *segment
		// positions in which this segment repeats itself, max 32 positions.
		pos []uint32 // todo: no need to hold pos if there is only 1 value.
	}

	diskSegment struct {
		*segment

		// order represents in which order this segment should be executed on decode.
		// max 255 segments per block.
		// TODO: maybe allow dynamic sizing here as well.
		order []uint8
	}
)

func (s *segment) Decompress() []byte {
	switch s.flags.getType() {
	case typeUncompressed:
		return s.buffer
	case typeRepeat:
		return bytes.Repeat(s.buffer, int(s.repeat))
	default:
		panic("invalid segment type")
	}
}

// GetCompressionGains returns how many bytes this section is compressing.
func (s *segment) GetCompressionGains() int64 {
	repeat := int64(s.repeat)
	bufLen := int64(len(s.buffer))
	posLen := int64(len(s.pos))

	originalSize := repeat * bufLen * posLen

	// meta 	= 1 byte
	// bufLen = 4 bytes
	// -------------------
	// total	=	5 bytes
	compressedSize := int64(5)
	// if segment is repeat, then +1 or +2 bytes.
	if s.flags.getType() == typeRepeat {
		compressedSize += 1
		if s.flags.isRepeat2Bytes() {
			compressedSize += 1
		}
	}
	compressedSize += posLen
	compressedSize += bufLen
	gain := originalSize - compressedSize
	return gain
}

// GetOrderedSegments will convert a Segment into an OrderedSegment,
// we do this to map pos, which occupies 4 bytes, to order, which is uses 1 byte.
func GetOrderedSegments(s *segment) []diskSegment {
	// segmentProjection is a projection abstraction to convert uint32 system coordinate to uint8.
	// we can run this optimization because we don't care about segment position in final buffer,
	// we only care about in which order they are decompressed.
	type segmentProjection struct {
		pos     uint32
		order   uint8
		segment *diskSegment
	}
	var projections []segmentProjection
	var segList []*diskSegment
	cur := s
	for {
		curSegment := &diskSegment{
			segment: cur,
		}
		segList = append(segList, curSegment)
		for _, pos := range cur.pos {
			projections = append(projections, segmentProjection{
				pos:     pos,
				segment: curSegment,
			})
		}
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	sort.Slice(projections, func(i, j int) bool {
		return projections[i].pos < projections[j].pos
	})
	// add a crescent order for all segments. example: 0,1,2,3.
	for i := 1; i < len(projections); i++ {
		order := projections[i-1].order + 1
		if order == 0 {
			panic("segment count overflow")
		}
		projections[i].order = order
	}
	// all projections link to their respective segments, so the same segment can have many projections.
	for _, entry := range projections {
		entry.segment.order = append(entry.segment.order, entry.order)
	}
	// here we are simply converting the projection back to an orderedSegment.
	finalList := make([]diskSegment, 0, len(segList))
	for i := range segList {
		finalList = append(finalList, *segList[i])
	}
	return finalList
}

// NewSegment creates a new segment.
func NewSegment(t segType, pos uint32, repeat uint16, buffer []byte) *segment {
	flags := meta(0)
	flags = flags.setIsRepeat2Bytes(repeat > math.MaxUint8)
	flags = flags.setPosLen(1)
	flags = flags.setType(t)
	resp := &segment{
		flags:  flags,
		repeat: repeat,
		buffer: buffer,
		pos:    []uint32{pos},
	}
	return resp
}

// AddPos will append all positions from pos into the current segment,
// it will return error if it overflows the maximum capacity of the segment.
func (s *segment) AddPos(pos []uint32) (*segment, error) {
	newLen := len(s.pos) + len(pos)
	if newLen > 0b11111 {
		// TODO: add better handling for repeating groups.
		return s, fmt.Errorf("len(pos) overflow")
	}
	s.flags = s.flags.setPosLen(uint8(newLen))
	s.pos = append(s.pos, pos...)
	return s, nil
}

// Add append a new segment to the linked list.
func (s *segment) Add(next *segment) *segment {
	s.next = next
	next.previous = s
	return next
}

// Remove dereferences this segment from the linked list.
func (s *segment) Remove() {
	if s.previous != nil {
		s.previous.next = s.next
	}
	if s.next != nil {
		s.next.previous = s.previous
	}
}

// Deduplicate will find segments that are identical, besides position, and merge them.
func (s *segment) Deduplicate() {
	cur := s
	for {
		iter := cur
		for {
			if iter = iter.next; iter == nil {
				break
			}
			if !bytes.Equal(cur.buffer, iter.buffer) || cur.repeat != iter.repeat || cur.flags.getType() != iter.flags.getType() {
				continue
			}
			// if pos doesn't overflow, we continue with the merge operation.
			if _, err := cur.AddPos(iter.pos); err == nil {
				iter.Remove()
			}
		}
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
}

// Optimize is responsible for finding segments that are causing byte compression gain to be negative, and try to
// revert it.
func (s *segment) Optimize() {
	cur := s
	for {
		if cur == nil {
			break
		}
		if cur.GetCompressionGains() > 0 || cur.previous == nil {
			cur = cur.next
			continue
		}
		type previousNextMap struct {
			prev, next *segment
		}
		siblings := make([]previousNextMap, 0, len(cur.pos))
		for _, pos := range cur.pos {
			if prev := s.FindSegment(pos - 1); prev != nil && prev.repeat == 1 && prev.flags.getType() == typeUncompressed {
				siblings = append(siblings, previousNextMap{
					prev: prev,
				})
				continue
			}
			if next := s.FindSegment(pos + 1); next != nil && next.repeat == 1 && next.flags.getType() == typeUncompressed {
				siblings = append(siblings, previousNextMap{
					next: next,
				})
			}
		}
		// Check if all possible positions are revertable.
		if len(siblings) == len(cur.pos) {
			for _, entry := range siblings {
				buf := bytes.Repeat(cur.buffer, int(cur.repeat))
				if entry.prev != nil {
					entry.prev.buffer = append(entry.prev.buffer, buf...)
					continue
				}
				if entry.next != nil {
					// since we are tweaking the beginning of the buffer, we have to update position.
					entry.next.pos[0] -= uint32(len(buf))
					entry.next.buffer = append(buf, entry.next.buffer...)
				}
			}
			cur.Remove()
		}
		cur = cur.next
	}
}

// FindSegment finds the segment that contains pos.
func (s *segment) FindSegment(pos uint32) *segment {
	cur := s
	for {
		for _, curPos := range cur.pos {
			if curPos <= pos && curPos+uint32(cur.repeat)*uint32(len(cur.buffer)) > pos {
				return s
			}
		}
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	return nil
}
