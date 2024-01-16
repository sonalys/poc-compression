package gompressor

import (
	"encoding/binary"
	"math"
)

var encoder = binary.BigEndian
var decoder = binary.BigEndian

func (cur *Segment) Encode() []byte {
	bufLen := int64(len(cur.Buffer))
	posLen := int64(len(cur.Pos))
	// allocate buffers.
	buffer := make([]byte, 0, cur.GetCompressedSize())
	repeatSize := uint8(0)
	if cur.Repeat > math.MaxUint8 {
		repeatSize = 1
	}
	lenPosSize := uint8(0)
	if posLen > math.MaxUint8 {
		lenPosSize = 1
	}
	lenBufSize := uint8(0)
	switch {
	case bufLen > math.MaxUint32:
		lenBufSize = 3
	case bufLen > math.MaxUint16:
		lenBufSize = 2
	case bufLen > math.MaxUint8:
		lenBufSize = 1
	}
	maxPos := cur.MaxPos
	posSize := uint8(0)
	switch {
	case maxPos > math.MaxUint32:
		posSize = 3
	case maxPos > math.MaxUint16:
		posSize = 2
	case maxPos > math.MaxUint8:
		posSize = 1
	}
	// start storing the binary.
	meta := NewMetadata().
		Set(SegmentTypeMask, byte(cur.Type)).
		Set(RepeatSizeMask, repeatSize).
		Set(LenPosSizeMask, lenPosSize).
		Set(LenBufSizeMask, lenBufSize).
		Set(PosSizeMask, posSize)
	buffer = append(buffer, meta.ToByte())
	if cur.Type == TypeRepeatSameChar {
		if cur.Repeat > math.MaxUint8 {
			buffer = encoder.AppendUint16(buffer, cur.Repeat)
		} else {
			buffer = append(buffer, byte(cur.Repeat))
		}
	}
	switch lenPosSize {
	case 0:
		buffer = append(buffer, byte(posLen))
	case 1:
		buffer = encoder.AppendUint16(buffer, uint16(posLen))
	}
	// we don't need to store the first position, since our decompression logic doesn't use it.
	for i := range cur.Pos {
		switch posSize {
		case 0:
			buffer = append(buffer, byte(cur.Pos[i]))
		case 1:
			buffer = encoder.AppendUint16(buffer, uint16(cur.Pos[i]))
		case 2:
			buffer = encoder.AppendUint32(buffer, uint32(cur.Pos[i]))
		case 3:
			buffer = encoder.AppendUint64(buffer, uint64(cur.Pos[i]))
		}
	}
	switch lenBufSize {
	case 0:
		buffer = append(buffer, byte(bufLen))
	case 1:
		buffer = encoder.AppendUint16(buffer, uint16(bufLen))
	case 2:
		buffer = encoder.AppendUint32(buffer, uint32(bufLen))
	case 3:
		buffer = encoder.AppendUint64(buffer, uint64(bufLen))
	}
	buffer = append(buffer, cur.Buffer...)
	return buffer
}

func DecodeSegment(b []byte) (*Segment, int64) {
	var pos int64
	flag, pos := Metadata(b[pos]), pos+1
	cur := Segment{
		Type:   SegmentType(flag.Check(SegmentTypeMask)),
		Repeat: 1,
	}
	if cur.Type == TypeRepeatSameChar {
		switch flag.Check(RepeatSizeMask) {
		case 0:
			cur.Repeat, pos = uint16(b[pos]), pos+1
		case 1:
			cur.Repeat, pos = decoder.Uint16(b[pos:]), pos+2
		}
	}
	var posLen int16
	switch flag.Check(LenPosSizeMask) {
	case 0:
		posLen, pos = int16(b[pos]), pos+1
	case 1:
		posLen, pos = int16(decoder.Uint16(b[pos:])), pos+2
	}
	cur.Pos = make([]int64, posLen)
	for i := range cur.Pos {
		switch flag.Check(PosSizeMask) {
		case 0:
			cur.Pos[i], pos = int64(b[pos]), pos+1
		case 1:
			cur.Pos[i], pos = int64(decoder.Uint16(b[pos:])), pos+2
		case 2:
			cur.Pos[i], pos = int64(decoder.Uint32(b[pos:])), pos+4
		case 3:
			cur.Pos[i], pos = int64(decoder.Uint64(b[pos:])), pos+8
		}
	}
	var bufLen int64
	switch flag.Check(LenBufSizeMask) {
	case 0:
		bufLen, pos = int64(b[pos]), pos+1
	case 1:
		bufLen, pos = int64(decoder.Uint16(b[pos:])), pos+2
	case 2:
		bufLen, pos = int64(decoder.Uint32(b[pos:])), pos+4
	case 3:
		bufLen, pos = int64(decoder.Uint64(b[pos:])), pos+8
	}
	cur.Buffer = b[pos : pos+bufLen]
	return &cur, pos + bufLen
}

func Compress(b *Block) []byte {
	out := make([]byte, 0, 8+len(b.Buffer))
	out = encoder.AppendUint64(out, uint64(b.Size))
	out = encoder.AppendUint64(out, uint64(len(b.Buffer)))
	out = append(out, b.Buffer...)
	if b.List == nil {
		return out
	}
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

func Decode(b []byte) (out *Block, err error) {
	var pos int64
	out = &Block{
		Size: int64(decoder.Uint64(b[0:])),
		List: &LinkedList[Segment]{},
	}
	pos += 8
	out.Buffer = b[pos : pos+int64(decoder.Uint64(b[4:]))]
	pos += int64(len(out.Buffer))
	cur := out.List
	for {
		if pos == int64(len(b)) {
			break
		}
		decoded, offset := DecodeSegment(b[pos:])
		cur.AppendValue(decoded)
		pos += offset
	}
	return
}
