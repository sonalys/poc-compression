package gompressor

import "encoding/binary"

var encoder = binary.BigEndian
var decoder = binary.BigEndian

func (cur orderedSegment) encodeSegment() []byte {
	bufLen := uint32(len(cur.buffer))
	// allocate buffers.
	orderLen := uint8(len(cur.order))
	cur.flags = cur.flags.setPosLen(orderLen)
	buffer := make([]byte, 0, 7+bufLen+uint32(orderLen))
	// start storing the binary.
	buffer = append(buffer, byte(cur.flags))
	buffer = encoder.AppendUint32(buffer, bufLen)
	if cur.flags.getType() == typeRepeat {
		if cur.flags.isRepeat2Bytes() {
			buffer = encoder.AppendUint16(buffer, cur.repeat)
		} else {
			buffer = append(buffer, byte(cur.repeat))
		}
	}
	// we don't need to store the first position, since our decompression logic doesn't use it.
	for i := range cur.order {
		buffer = append(buffer, byte(cur.order[i]))
	}
	buffer = append(buffer, cur.buffer...)
	return buffer
}

func decodeSegment(b []byte) (orderedSegment, uint32) {
	var pos uint32
	flag := meta(b[pos])
	pos += 1
	cur := orderedSegment{
		segment: &segment{
			flags:  flag,
			repeat: 1,
		},
		order: make([]byte, flag.getPosLen()),
	}
	bufLen := decoder.Uint32(b[pos:])
	pos += 4
	cur.buffer = make([]byte, bufLen)
	if flag.getType() == typeRepeat {
		if flag.isRepeat2Bytes() {
			cur.repeat = decoder.Uint16(b[pos:])
			pos += 2
		} else {
			cur.repeat = uint16(b[pos])
			pos += 1
		}
	}
	for i := range cur.order {
		cur.order[i] = b[pos]
		pos += 1
	}
	cur.buffer = b[pos : pos+bufLen]
	return cur, pos + bufLen
}

func encode(b *block) []byte {
	buffer := make([]byte, 0, b.size)
	// Store original size of the buffer.
	buffer = encoder.AppendUint32(buffer, b.size)
	// Iterate from head to tail of segments.
	for _, entry := range b.head {
		buffer = append(buffer, entry.encodeSegment()...)
	}
	return buffer
}

func decode(b []byte) (out *block, err error) {
	var pos uint32
	out = &block{
		parsed: true,
		size:   decoder.Uint32(b[0:]),
		head:   make([]orderedSegment, 0),
	}
	pos += 4
	var i uint8
	for {
		if pos == uint32(len(b)) {
			break
		}
		cur, offset := decodeSegment(b[pos:])
		out.head = append(out.head, cur)
		pos += offset
		i++
		if i == 0 {
			panic("segment overflow")
		}
	}
	return
}
