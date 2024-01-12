package gompressor

import (
	"encoding/binary"
	"math"
)

var encoder = binary.BigEndian
var decoder = binary.BigEndian

func encodeValue[S BlockSize](b []byte, value S) []byte {
	switch v := any(value).(type) {
	case uint8:
		return append(b, v)
	case uint16:
		return encoder.AppendUint16(b, v)
	case uint32:
		return encoder.AppendUint32(b, v)
	case uint64:
		return encoder.AppendUint64(b, v)
	default:
		panic("invalid blockSize type")
	}
}

func decodeValue[S BlockSize](b []byte) (resp S) {
	switch v := any(resp).(type) {
	case uint8:
		v = uint8(b[0])
		return S(v)
	case uint16:
		v = decoder.Uint16(b)
		return S(v)
	case uint32:
		v = decoder.Uint32(b)
		return S(v)
	case uint64:
		v = decoder.Uint64(b)
		return S(v)
	default:
		panic("invalid blockSize type")
	}
}

func (cur *Segment[S]) Encode() []byte {
	bufLen := S(len(cur.Buffer))
	// allocate buffers.
	posLen := uint8(len(cur.Pos))
	buffer := make([]byte, 0, 7+bufLen+S(posLen))
	// start storing the binary.
	buffer = append(buffer, byte(NewMetadata(cur.Type, posLen, cur.Repeat > math.MaxUint8)))
	buffer = encodeValue(buffer, bufLen)
	if cur.Type == TypeRepeatSameChar {
		if cur.Repeat > math.MaxUint8 {
			buffer = encoder.AppendUint16(buffer, cur.Repeat)
		} else {
			buffer = append(buffer, byte(cur.Repeat))
		}
	}
	// we don't need to store the first position, since our decompression logic doesn't use it.
	for i := range cur.Pos {
		buffer = encodeValue(buffer, cur.Pos[i])
	}
	buffer = append(buffer, cur.Buffer...)
	return buffer
}

func DecodeSegment[S BlockSize](b []byte) (*Segment[S], S) {
	var pos S
	flag := meta(b[pos])
	pos += 1
	cur := Segment[S]{
		Type:   flag.getType(),
		Repeat: 1,
		Pos:    make([]S, flag.getPosLen()),
	}
	bufLen := decodeValue[S](b[pos:])
	pos += 4
	cur.Buffer = make([]byte, bufLen)
	if flag.getType() == TypeRepeatSameChar {
		if flag.isRepeat2Bytes() {
			cur.Repeat = decoder.Uint16(b[pos:])
			pos += 2
		} else {
			cur.Repeat = uint16(b[pos])
			pos += 1
		}
	}
	for i := range cur.Pos {
		cur.Pos[i] = decodeValue[S](b[pos:])
		pos += 4
	}
	cur.Buffer = b[pos : pos+bufLen]
	return &cur, pos + bufLen
}

func Encode[S BlockSize](b *Block[S]) []byte {
	out := make([]byte, 0, 8+len(b.Buffer))
	// Store original size of the buffer.
	out = encodeValue(out, b.Size)
	out = encodeValue(out, S(len(b.Buffer)))
	out = append(out, b.Buffer...)
	cur := b.List.Head
	for {
		if cur == nil {
			break
		}
		out = append(out, cur.Value.Encode()...)
		cur = cur.Next
	}
	return out
}

func Decode[S BlockSize](b []byte) (out *Block[S], err error) {
	var pos S
	out = &Block[S]{
		Size: decodeValue[S](b[0:]),
		List: &LinkedList[Segment[S]]{},
	}
	pos += 8
	out.Buffer = b[pos : pos+decodeValue[S](b[4:])]
	pos += S(len(out.Buffer))
	cur := out.List
	for {
		if pos == S(len(b)) {
			break
		}
		decoded, offset := DecodeSegment[S](b[pos:])
		cur.AppendValue(decoded)
		pos += offset
	}
	return
}
