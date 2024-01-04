package gompressor

import (
	"encoding/binary"
	"math"
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
		size uint32
		head *segment
	}

	// segment
	// memory serialization requires magic to avoid storing as much bytes as possible,
	// while still being able to recover the struct and reconstruct the original buffer.
	segment struct {
		flags meta
		// pos indicates where the buffer is repeating in the file, it can repeat identically in many places.
		// we can have at max 32 positions per segment.
		pos []uint32
		// repeat indicates how many times buffer is repeated. that's not the same as pos.
		// we are only storing this field in disk if the segment type is typeRepeat.
		repeat uint16
		// buffer can hold at maximum 4294967295 bytes, or 4.294967 gigabytes.
		buffer []byte
		// no need to serialize these fields on disk.
		previous, next *segment
	}
)

const (
	typeUncompressed segType = 0b0
	typeRepeat       segType = 0b1

	flagRepeatIs2Bytes meta = 0b1 << 2
)

// setPosLen
// 1. clears the last 5 bytes
// 2. left shift value 3 bytes
// 3. set value of posLen.
func (m meta) setPosLen(size byte) meta {
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

func (m meta) IsRepeat2Bytes() bool { return m&flagRepeatIs2Bytes != 0 }

func newSegment(t segType, pos uint32, repeat uint16, buffer []byte) *segment {
	flags := meta(0)
	if repeat > math.MaxUint8 {
		flags = flags | flagRepeatIs2Bytes
	}
	flags = flags.setPosLen(1)
	flags = flags.setType(t)
	return &segment{
		flags:  flags,
		repeat: repeat,
		buffer: buffer,
		pos:    []uint32{pos},
	}
}

func (s *segment) addNext(t segType, pos uint32, repeat uint16, buffer []byte) *segment {
	s.next = newSegment(t, pos, repeat, buffer)
	s.next.previous = s
	return s.next
}

func (s *segment) addPos(pos []uint32) *segment {
	newLen := s.flags.getPosLen() + byte(len(pos))
	if len(pos) > 31 {
		// TODO: add better handling for repeating groups.
		panic("len(pos) overflow")
	}
	s.flags = s.flags.setPosLen(newLen)
	s.pos = append(s.pos, pos...)
	return s
}

var encoder = binary.BigEndian

func (cur segment) serialize(i uint32) []byte {
	buffer := make([]byte, 0, len(cur.buffer))
	buffer = append(buffer, byte(cur.flags))
	buffer = encoder.AppendUint32(buffer, uint32(len(cur.buffer)))
	if cur.flags.IsRepeat2Bytes() {
		buffer = encoder.AppendUint16(buffer, cur.repeat)
	} else {
		buffer = append(buffer, byte(cur.repeat))
	}
	for i := range cur.pos {
		buffer = encoder.AppendUint32(buffer, cur.pos[i])
	}
	buffer = append(buffer, cur.buffer...)
	return buffer
}

func (b *block) serialize() []byte {
	buffer := make([]byte, 0, b.size)
	// Store original size of the buffer.
	buffer = encoder.AppendUint32(buffer, b.size)
	// Iterate from head to tail of segments.
	cur := b.head
	var i uint32
	for {
		buffer = append(buffer, cur.serialize(i)...)
		if cur.next == nil {
			break
		}
		cur = cur.next
		i++
	}
	return buffer
}

// deduplicate repetition groups of same size and same buffer.
func (b *segment) deduplicate() {
	segI := b
	for {
		segJ := segI
		for {
			segJ = segJ.next
			if segJ == nil {
				break
			}
			if segI.flags.getType() != segJ.flags.getType() || // check segment type for merging.
				segI.buffer[0] == segJ.buffer[0] || // check if the repeated char is the same.
				segI.repeat != segJ.repeat { // check if repeat is the same amount.
				continue
			}
			segI.addPos(segJ.pos)
			segJ.previous.next = segJ.next // delete this element
		}
		if segI.next == nil {
			break
		}
		segI = segI.next
	}
}
