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
		Size     uint32
		Segments []DiskSegment
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

	DiskSegment struct {
		*Segment

		// Order represents in which Order this segment should be executed on decode.
		// max 255 segments per block.
		// TODO: maybe allow dynamic sizing here as well.
		Order []uint16
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
	compressedSize += posLen * 2
	compressedSize += bufLen
	gain := originalSize - compressedSize
	return gain
}

// GetOrderedSegments will convert a Segment into an OrderedSegment,
// we do this to map pos, which occupies 4 bytes, to order, which is uses 1 byte.
func GetOrderedSegments(head *Segment) []DiskSegment {
	// segmentProjection is a projection abstraction to convert uint32 system coordinate to uint8.
	// we can run this optimization because we don't care about segment position in final buffer,
	// we only care about in which order they are decompressed.
	type segmentProjection struct {
		pos     uint32
		order   uint16
		segment *DiskSegment
	}
	var projections []segmentProjection
	var segList []*DiskSegment
	cur := head
	for {
		curSegment := &DiskSegment{
			Segment: cur,
		}
		segList = append(segList, curSegment)
		for _, pos := range cur.Pos {
			projections = append(projections, segmentProjection{
				pos:     pos,
				segment: curSegment,
			})
		}
		if cur.Next == nil {
			break
		}
		cur = cur.Next
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
		entry.segment.Order = append(entry.segment.Order, entry.order)
	}
	// here we are simply converting the projection back to an orderedSegment.
	finalList := make([]DiskSegment, 0, len(segList))
	for i := range segList {
		finalList = append(finalList, *segList[i])
	}
	return finalList
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
func (s *Segment) Remove() {
	if s.Previous != nil {
		s.Previous.Next = s.Next
	}
	if s.Next != nil {
		s.Next.Previous = s.Previous
	}
}

// Deduplicate will find segments that are identical, besides position, and merge them.
func (s *Segment) Deduplicate() {
	cur := s
	for {
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
				iter.Remove()
			}
		}
		if cur.Next == nil {
			break
		}
		cur = cur.Next
	}
}

// IsMergeable returns true if the segment can be merged with another.
// TODO: maybe improve this logic, and receive another segment s2,
// Calculate if they are always together relative to each other, but might not be worth the effort.
func (s *Segment) IsMergeable() bool {
	// For a segment to be mergeable, you need for it to only appear once,
	// be uncompressed, and without repetitions.
	// Otherwise you might merge segments of different repeats and positions, and distort the final data.
	return len(s.Pos) == 1 && s.Repeat == 1 && s.Metadata.getType() == typeUncompressed
}

// Optimize is responsible for finding segments that are causing byte compression gain to be negative, and try to
// revert it.
func (s *Segment) Optimize() {
	cur := s
	for {
		if cur == nil {
			break
		}
		if cur.GetCompressionGains() > 0 || cur.Previous == nil {
			cur = cur.Next
			continue
		}
		type previousNextMap struct {
			prev, next *Segment
		}
		mergeable := make([]previousNextMap, 0, len(cur.Pos))
		for _, pos := range cur.Pos {
			if sibling, found := s.FindSegment(pos - 1); found && sibling.IsMergeable() {
				mergeable = append(mergeable, previousNextMap{
					prev: sibling,
				})
				continue
			}
			if sibling, found := s.FindSegment(pos + uint32(len(cur.Buffer))); found && sibling.IsMergeable() {
				mergeable = append(mergeable, previousNextMap{
					next: sibling,
				})
			}
		}
		// Check if all possible positions are mergeable.
		if len(mergeable) == len(cur.Pos) {
			for _, entry := range mergeable {
				buf := bytes.Repeat(cur.Buffer, int(cur.Repeat))
				if entry.prev != nil {
					entry.prev.Buffer = append(entry.prev.Buffer, buf...)
					continue
				}
				if entry.next != nil {
					// since we are tweaking the beginning of the buffer, we have to update position.
					entry.next.Pos[0] -= uint32(len(buf))
					entry.next.Buffer = append(buf, entry.next.Buffer...)
				}
			}
			cur.Remove()
		}
		cur = cur.Next
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
