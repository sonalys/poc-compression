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
		flags meta
		// pos indicates where the buffer is repeating in the file, it can repeat identically in many places.
		// we can have at max 32 positions per segment.
		pos []uint32 // todo: no need to hold pos if there is only 1 value.
		// repeat indicates how many times buffer is repeated. that's not the same as pos.
		// we are only storing this field in disk if the segment type is typeRepeat.
		repeat uint16
		// buffer can hold at maximum 4294967295 bytes, or 4.294967 gigabytes.
		buffer []byte
		// no need to serialize these fields on disk.
		previous, next *segment
	}

	orderedSegment struct {
		*segment
		order []uint8
	}
)

// returns how many bytes this section is compressing.
func (s *segment) getCompressionGains() int {
	var compressedSize int = 6 // meta + bufLen + 1 byte repeat
	if s.flags.isRepeat2Bytes() {
		compressedSize += 1 // +1 byte repeat
	}
	compressedSize += int(s.flags.getPosLen()) // 1 byte per order
	compressedSize += int(len(s.buffer))
	originalSize := int(len(s.buffer)) * int(s.repeat) * int(len(s.pos))
	return originalSize - compressedSize
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
	var list []segmentProjection
	cur := s
	for {
		curSegment := &orderedSegment{
			segment: cur,
		}
		for _, pos := range cur.pos {
			list = append(list, segmentProjection{
				pos:     pos,
				segment: curSegment,
			})
		}
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].pos < list[j].pos
	})

	for i := 1; i < len(list); i++ {
		list[i].order = list[i-1].order + uint8(1)
	}
	orderedSegmentList := make([]orderedSegment, 0, len(list))
	for _, entry := range list {
		entry.segment.order = append(entry.segment.order, entry.order)
	}
	for _, entry := range list {
		orderedSegmentList = append(orderedSegmentList, *entry.segment)
	}
	return orderedSegmentList
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

func newSegment(t segType, pos uint32, repeat uint16, buffer []byte) (*segment, int) {
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
	gains := resp.getCompressionGains()
	return resp, gains
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
