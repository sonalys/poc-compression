package main

import (
	"bytes"
	"encoding/binary"
)

// segment
// memory serialization requires magic to avoid storing as much bytes as possible,
// while still being able to recover the struct and reconstruct the original buffer.
// I am thinking in doing the following:
// Address	|	[file.gompressed]	|	size
// 0x0000			block.size					4 bytes
// SEGMENT SECTION [ 5 + 4 * pos + data bytes ]
// +4					segment.repeat			2 bytes
// +2					segment.pos.len			1 byte
// +1					segment.buf.len			4 bytes
// +4		 			segment.pos					LEN * 4 bytes
// +LEN				segment.buf					LEN bytes
type segment struct {
	positions []uint32
	repeat    uint16

	buffer []byte
	// no need to serialize these fields on disk.
	previous, next *segment
}

type block struct {
	size uint32
	head *segment
}

func (b *block) serialize() []byte {
	buffer := make([]byte, 0, b.size)
	encoder := binary.BigEndian

	buffer = encoder.AppendUint32(buffer, b.size)
	cur := b.head
	for {
		buffer = encoder.AppendUint16(buffer, cur.repeat)
		buffer = append(buffer, byte(len(cur.positions)))
		buffer = encoder.AppendUint32(buffer, uint32(len(cur.buffer)))
		for i := range cur.positions {
			buffer = encoder.AppendUint32(buffer, cur.positions[i])
		}
		buffer = append(buffer, cur.buffer...)
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	return buffer
}

func parse(b []byte) (out block, err error) {
	decoder := binary.BigEndian
	pos := 0
	out.size = decoder.Uint32(b[pos:])
	pos += 4
	out.head = &segment{}
	cur := out.head
	for {
		if pos == len(b) {
			break
		}
		temp := segment{
			repeat:    uint16(decoder.Uint16(b[pos:])),
			positions: make([]uint32, b[pos+2]),
			buffer:    make([]byte, decoder.Uint32(b[pos+3:])),
		}
		pos += 7
		for i := range temp.positions {
			temp.positions[i] = decoder.Uint32(b[pos:])
			pos += 4
		}
		temp.buffer = b[pos : pos+len(temp.buffer)]
		pos += len(temp.buffer)
		cur.next = &temp
		cur = cur.next
	}
	out.head = out.head.next
	return
}

// deduplicate repetition groups of same size and same buffer.
func deduplicate(s *segment) {
	cur := s
	for {
		if cur.next != nil {
			break
		}
		next := cur.next
		if cur.repeat != next.repeat || cur.buffer[0] == next.buffer[0] {
			continue
		}
		cur.positions = append(cur.positions, next.positions...)
		next.previous.next = next.next // delete this element
		cur = cur.next
	}
}

func compress(in []byte, minSize uint16) block {
	b := block{
		size: uint32(len(in)),
		head: &segment{},
	}
	// prev is a cursor to the last position before a repeating group
	var prev uint32
	// cur is a cursor to the head of the block's segments.
	cur := b.head
	// finds repetition groups and store them.
	for i := uint32(0); i < uint32(len(in)); i++ {
		size := uint16(1)
		for j := i + 1; j < uint32(len(in)) && in[i] == in[j]; j++ {
			size++
		}
		if size > minSize {
			cur.next = &segment{
				positions: []uint32{prev},
				buffer:    in[prev:i],
				repeat:    1,
				previous:  cur,
			}
			cur = cur.next
			cur.next = &segment{
				positions: []uint32{i},
				buffer:    []byte{in[i]},
				repeat:    size,
				previous:  cur,
			}
			cur = cur.next
			i += uint32(size) - 1
			prev = i
		}
	}
	if b.head.next != nil {
		b.head = b.head.next
	}
	// append last segment of non-repeting characters.
	cur.next = &segment{
		positions: []uint32{prev},
		buffer:    in[prev:],
		repeat:    1,
		previous:  cur,
	}

	deduplicate(b.head)

	// for i := uint32(0); i < uint32(len(out)); i++ {
	// 	localGroups := []repetitionGroup{}
	// 	maxSize := uint32(0)
	// outer:
	// 	for j := i + minSize; j < uint32(len(in)); j++ {
	// 		size := uint32(minSize)
	// 		// finds the minimum size that matches i and j.
	// 		for ; size < j-i; size++ {
	// 			if out[i+size] != out[j+size] {
	// 				break
	// 			}
	// 		}
	// 		if size < minSize {
	// 			continue
	// 		}
	// 		if size > maxSize {
	// 			maxSize = size
	// 		}
	// 		for k := range groups {
	// 			if bytes.Equal(groups[k].Bytes, out[i:i+size]) {
	// 				// already registered previously.
	// 				continue outer
	// 			}
	// 		}
	// 		// if there is already a group for i with same size, then bytes is equal as well.
	// 		for k := range localGroups {
	// 			if uint32(len(localGroups[k].Bytes)) == size {
	// 				localGroups[k].Positions = append(localGroups[k].Positions, j)
	// 			}
	// 			continue outer
	// 		}
	// 		// if there is no local group, then create one for i, j.
	// 		localGroups = append(localGroups, repetitionGroup{
	// 			Positions: []uint32{i, j},
	// 			Bytes:     out[i : i+size],
	// 		})
	// 	}
	// 	if len(localGroups) == 0 {
	// 		out2 = append(out2, out[i])
	// 		continue
	// 	}
	// 	i += maxSize
	// 	groups = append(groups, localGroups...)
	// }
	// out = out2
	return b
}

func reconstruct(in block) []byte {
	out := make([]byte, in.size)
	cur := in.head
	for {
		from := bytes.Repeat(cur.buffer, int(cur.repeat))
		for _, pos := range cur.positions {
			copy(out[pos:], from)
		}
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	return out
}

// func reconstruct(in block) (out []byte) {
// 	out = make([]byte, 0, len(in))
// 	prevPos := 0
// 	for _, entry := range entries {
// 		curPos := int(entry.Position)
// 		out = append(out, in[prevPos:curPos]...)
// 		out = append(out, bytes.Repeat([]byte{entry.Byte}, int(entry.Size))...)
// 		prevPos = curPos
// 	}
// 	out = append(out, in[prevPos:]...)
// 	return
// }

func countBytes(input []byte) (repetition []uint) {
	repetition = make([]uint, 256)
	for _, char := range input {
		repetition[char] = repetition[char] + 1
	}
	return
}
