package gompressor

import (
	"bytes"
	"encoding/binary"
)

var decoder = binary.BigEndian

func parseSegment(b []byte, i uint32) (segment, uint32) {
	var pos uint32
	flag := meta(b[pos])
	pos += 1
	seg := segment{
		flags: flag,
		pos:   make([]uint32, flag.getPosLen()),
	}
	seg.buffer = make([]byte, decoder.Uint32(b[pos:]))
	pos += 4
	if flag.IsRepeat2Bytes() {
		seg.repeat = decoder.Uint16(b[pos:])
		pos += 2
	} else {
		seg.repeat = uint16(b[pos])
		pos += 1
	}
	for i := range seg.pos {
		seg.pos[i] = decoder.Uint32(b[pos:])
		pos += 4
	}
	bufLen := uint32(len(seg.buffer))
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

func decompress(in block) []byte {
	out := make([]byte, in.size)
	cur := in.head
	for {
		from := bytes.Repeat(cur.buffer, int(cur.repeat))
		for _, pos := range cur.pos {
			copy(out[pos:], from)
		}
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	return out
}
