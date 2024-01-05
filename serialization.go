package gompressor

import "encoding/binary"

var encoder = binary.BigEndian

func (cur segment) serialize(i uint32) []byte {
	bufLen := uint32(len(cur.buffer))
	posLen := uint32(len(cur.pos))

	buffer := make([]byte, 0, 7+bufLen+4*posLen)

	buffer = append(buffer, byte(cur.flags))
	buffer = encoder.AppendUint32(buffer, bufLen)
	if cur.flags.getType() == typeRepeat {
		if cur.flags.isRepeat2Bytes() {
			buffer = encoder.AppendUint16(buffer, cur.repeat)
		} else {
			buffer = append(buffer, byte(cur.repeat))
		}
	}
	for i := range cur.pos {
		buffer = encoder.AppendUint32(buffer, cur.pos[i])
	}
	buffer = append(buffer, cur.buffer...)
	return buffer
}

func (b block) serialize() []byte {
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

var decoder = binary.BigEndian

func parseSegment(b []byte, i uint32) (segment, uint32) {
	var pos uint32
	flag := meta(b[pos])
	pos += 1
	seg := segment{
		flags: flag,
		pos:   make([]uint32, flag.getPosLen()),
	}
	bufLen := decoder.Uint32(b[pos:])
	pos += 4
	seg.buffer = make([]byte, bufLen)
	if flag.getType() == typeRepeat {
		if flag.isRepeat2Bytes() {
			seg.repeat = decoder.Uint16(b[pos:])
			pos += 2
		} else {
			seg.repeat = uint16(b[pos])
			pos += 1
		}
	}
	for i := range seg.pos {
		seg.pos[i] = decoder.Uint32(b[pos:])
		pos += 4
	}
	seg.buffer = b[pos : pos+bufLen]
	return seg, pos + bufLen
}

func parse(b []byte) (out block, err error) {
	var pos uint32
	out.size = decoder.Uint32(b[0:])
	pos += 4
	out.head = &segment{}
	cur := out.head
	var i uint32
	for {
		if pos == uint32(len(b)) {
			break
		}
		seg, offset := parseSegment(b[pos:], i)
		pos += offset
		cur.next = &seg
		seg.previous = cur
		cur = cur.next
		i++
	}
	out.head = out.head.next
	return
}
