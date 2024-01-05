package gompressor

import (
	"bytes"
	"fmt"
	"math"
	"sort"
)

type (
	// meta is a bitmask used to store metadata.
	// Address	|		data
	//
	//	1, 2					Segment types = max 4.
	//	3							Repeat size, 0 = 1 byte, 1 = 2 bytes.
	//	4,5,6,7,8			posLen = max 32.
	meta uint8

	segType uint8

	// My cats are still trying to reach to me about how to increase segments efficiency in disk, so stay tuned.

	block struct {
		size   uint32
		parsed bool
		head   []orderedSegment
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

	orderedSegment struct {
		*segment

		// order represents in which order this segment should be executed on decode.
		// max 255 segments per block.
		// TODO: maybe allow dynamic sizing here as well.
		order []uint8
	}
)

// returns how many bytes this section is compressing.
func (s *segment) getCompressionGains() int64 {
	repeat := int64(s.repeat)
	bufLen := int64(len(s.buffer))
	posLen := int64(len(s.pos))

	originalSize := repeat * bufLen * posLen

	// meta 	= 1 byte
	// repeat = 1 byte
	// bufLen = 4 bytes
	// -------------------
	// total	=	6 bytes
	compressedSize := int64(5)
	if s.flags.getType() == typeRepeat {
		compressedSize += 1
		if s.flags.isRepeat2Bytes() {
			// if repeat is 2 bytes, then we sum +1.
			compressedSize += 1
		}
	}
	// 1 byte per order, -1 because the first pos is discarded.
	compressedSize += posLen - 1
	compressedSize += bufLen
	gain := originalSize - compressedSize
	if gain < 0 {
		// print("fuck")
	}
	return gain
}

func getOrderedSegments(s *segment) []orderedSegment {
	// segmentProjection is a projection abstraction to convert uint32 system coordinate to uint8.
	// we can run this optimization because we don't care about segment position in final buffer,
	// we only care about in which order they are decompressed.
	type segmentProjection struct {
		pos     uint32
		order   uint8
		segment *orderedSegment
	}
	var projections []segmentProjection
	var segList []*orderedSegment
	cur := s
	for {
		curSegment := &orderedSegment{
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
		projections[i].order = projections[i-1].order + uint8(1)
	}
	// all projections link to their respective segments, so the same segment can have many projections.
	for _, entry := range projections {
		entry.segment.order = append(entry.segment.order, entry.order)
	}
	// here we are simply converting the projection back to an orderedSegment.
	finalList := make([]orderedSegment, 0, len(segList))
	for i := range segList {
		finalList = append(finalList, *segList[i])
	}
	return finalList
}

const (
	typeUncompressed segType = 0b0
	typeRepeat       segType = 0b1

	flagRepeatIs2Bytes meta = 0b1 << 2
)

// setPosLen
// 1. clears the last 5 bytes
// 2. left shift value 3 bytes
// 3. set value of posLen.
func (m meta) setPosLen(size uint8) meta {
	return m&0b111 | (meta(size) << 3)
}

// getPosLen
// right shift 3 bytes to get original posLen.
func (m meta) getPosLen() byte { return byte(m >> 3) }

// setType
// 1. clears bytes 2 and 3
// 1. set bytes 2 and 3
func (m meta) setType(t segType) meta { return (m & 0b11111100) | meta(t) }

// getType
// clear all bytes except 2 and 3
// shift right 1 byte to get segType
func (m meta) getType() segType { return segType(m & 0b11) }

func (m meta) isRepeat2Bytes() bool { return m&flagRepeatIs2Bytes != 0 }
func (m meta) setIsRepeat2Bytes(value bool) meta {
	if value {
		return m | flagRepeatIs2Bytes
	}
	return m &^ flagRepeatIs2Bytes
}

func newSegment(t segType, pos uint32, repeat uint16, buffer []byte) *segment {
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

func (s *segment) addPos(pos []uint32) (*segment, error) {
	newLen := len(s.pos) + len(pos)
	if newLen > 0b11111 {
		// TODO: add better handling for repeating groups.
		return s, fmt.Errorf("len(pos) overflow")
	}
	s.flags = s.flags.setPosLen(uint8(newLen))
	s.pos = append(s.pos, pos...)
	return s, nil
}

func (s *segment) add(next *segment) *segment {
	s.next = next
	next.previous = s
	return next
}

// deduplicate repetition groups of same size and same buffer.
func (b *segment) deduplicate() {
	current := b
	for {
		iter := current
		for {
			if iter = iter.next; iter == nil {
				break
			}
			if !bytes.Equal(current.buffer, iter.buffer) || current.repeat != iter.repeat || current.flags.getType() != iter.flags.getType() {
				continue
			}
			// if pos doesn't overflow, we continue with the merge operation.
			if _, err := current.addPos(iter.pos); err == nil {
				iter.previous.next = iter.next
			}
		}
		if current.next == nil {
			break
		}
		current = current.next
	}
}

func (b *segment) optimize() {
	cur := b
	modified := false
	for {
		if cur == nil {
			if modified {
				cur = b
				modified = false
				continue
			}
			break
		}
		if cur.getCompressionGains() > 0 || cur.previous == nil {
			cur = cur.next
			continue
		}
		previousList := make([]*segment, 0, len(cur.pos))
		for _, pos := range cur.pos {
			// If we don't find the previous segment in any block, we can't optimize it,
			// so we just go to the next optimizable segment.
			prev := b.findSegmentByPos(pos - 1)
			if prev == nil || prev.flags.getType() != typeUncompressed {
				break
			}
			previousList = append(previousList, prev)
		}
		// we batch them together to create an atomic operation, either we do it or we dont, no half-way.
		if len(previousList) == len(cur.pos) {
			for _, previous := range previousList {
				previous.buffer = append(previous.buffer, bytes.Repeat(cur.buffer, int(cur.repeat))...)
			}
			// remove segment from linked list.
			cur.previous.next = cur.next
			modified = true
		}
		cur = cur.next
	}
}

func (b *segment) findSegmentByPos(pos uint32) *segment {
	cur := b
	for {
		for _, curPos := range cur.pos {
			if curPos <= pos && curPos+uint32(cur.repeat)*uint32(len(cur.buffer)) > pos {
				return b
			}
		}
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	return nil
}
