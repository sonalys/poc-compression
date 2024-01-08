package gompressor

import "encoding/binary"

var encoder = binary.BigEndian
var decoder = binary.BigEndian

func (cur DiskSegment) Encode() []byte {
	bufLen := uint32(len(cur.Buffer))
	// allocate buffers.
	orderLen := uint8(len(cur.Order))
	cur.Metadata = cur.Metadata.setPosLen(orderLen)
	buffer := make([]byte, 0, 7+bufLen+uint32(orderLen))
	// start storing the binary.
	buffer = append(buffer, byte(cur.Metadata))
	buffer = encoder.AppendUint32(buffer, bufLen)
	if cur.Metadata.getType() == typeRepeat {
		if cur.Metadata.isRepeat2Bytes() {
			buffer = encoder.AppendUint16(buffer, cur.Repeat)
		} else {
			buffer = append(buffer, byte(cur.Repeat))
		}
	}
	// we don't need to store the first position, since our decompression logic doesn't use it.
	for i := range cur.Order {
		buffer = append(buffer, byte(cur.Order[i]))
	}
	buffer = append(buffer, cur.Buffer...)
	return buffer
}

func DecodeSegment(b []byte) (DiskSegment, uint32) {
	var pos uint32
	flag := meta(b[pos])
	pos += 1
	cur := DiskSegment{
		Segment: &Segment{
			Metadata: flag,
			Repeat:   1,
		},
		Order: make([]uint16, flag.getPosLen()),
	}
	bufLen := decoder.Uint32(b[pos:])
	pos += 4
	cur.Buffer = make([]byte, bufLen)
	if flag.getType() == typeRepeat {
		if flag.isRepeat2Bytes() {
			cur.Repeat = decoder.Uint16(b[pos:])
			pos += 2
		} else {
			cur.Repeat = uint16(b[pos])
			pos += 1
		}
	}
	for i := range cur.Order {
		cur.Order[i] = decoder.Uint16(b[pos:])
		pos += 2
	}
	cur.Buffer = b[pos : pos+bufLen]
	return cur, pos + bufLen
}

func Encode(b *block) []byte {
	buffer := make([]byte, 0, b.Size)
	// Store original size of the buffer.
	buffer = encoder.AppendUint32(buffer, b.Size)
	// Iterate from head to tail of segments.
	for _, entry := range b.Segments {
		buffer = append(buffer, entry.Encode()...)
	}
	return buffer
}

func Decode(b []byte) (out *block, err error) {
	var pos uint32
	out = &block{
		Size:     decoder.Uint32(b[0:]),
		Segments: make([]DiskSegment, 0),
	}
	pos += 4
	var i uint8
	for {
		if pos == uint32(len(b)) {
			break
		}
		cur, offset := DecodeSegment(b[pos:])
		out.Segments = append(out.Segments, cur)
		pos += offset
		i++
		if i == 0 {
			panic("segment overflow")
		}
	}
	return
}
