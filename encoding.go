package gompressor

import "encoding/binary"

var encoder = binary.BigEndian
var decoder = binary.BigEndian

func (cur *Segment) Encode() []byte {
	bufLen := uint32(len(cur.Buffer))
	// allocate buffers.
	posLen := uint8(len(cur.Pos))
	cur.Metadata = cur.Metadata.setPosLen(posLen)
	buffer := make([]byte, 0, 7+bufLen+uint32(posLen))
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
	for i := range cur.Pos {
		buffer = encoder.AppendUint32(buffer, cur.Pos[i])
	}
	buffer = append(buffer, cur.Buffer...)
	return buffer
}

func DecodeSegment(b []byte) (*Segment, uint32) {
	var pos uint32
	flag := meta(b[pos])
	pos += 1
	cur := Segment{
		Metadata: flag,
		Repeat:   1,
		Pos:      make([]uint32, flag.getPosLen()),
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
	for i := range cur.Pos {
		cur.Pos[i] = decoder.Uint32(b[pos:])
		pos += 4
	}
	cur.Buffer = b[pos : pos+bufLen]
	return &cur, pos + bufLen
}

func Encode(b *Block) []byte {
	out := make([]byte, 0, 8+len(b.Buffer))
	// Store original size of the buffer.
	out = encoder.AppendUint32(out, b.Size)
	out = encoder.AppendUint32(out, uint32(len(b.Buffer)))
	out = append(out, b.Buffer...)
	b.ForEach(func(s *Segment) {
		out = append(out, s.Encode()...)
	})
	return out
}

func Decode(b []byte) (out *Block, err error) {
	var pos uint32
	out = &Block{
		Size: decoder.Uint32(b[0:]),
		Head: &Segment{},
	}
	pos += 8
	out.Buffer = b[pos : pos+decoder.Uint32(b[4:])]
	pos += uint32(len(out.Buffer))
	cur := out.Head
	for {
		if pos == uint32(len(b)) {
			break
		}
		decoded, offset := DecodeSegment(b[pos:])
		cur = cur.Add(decoded)
		pos += offset
	}
	// Fix head.
	out.Remove(out.Head)
	return
}
