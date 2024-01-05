package gompressor

import (
	"bytes"
	"fmt"
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
		pos []uint32 // todo: no need to hold pos if there is only 1 value.
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
