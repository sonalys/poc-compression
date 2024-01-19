package gompressor

import "math"

func (s *Segment) Encode() []byte {
	bufLen := int(len(s.Buffer))
	posLen := int(len(s.Pos))
	// allocate buffers.
	buffer := make([]byte, 0, s.GetCompressedSize())
	meta := NewMetadata().SetType(s.Type)
	switch {
	case s.Repeat > math.MaxUint8:
		meta = meta.SetRepSize(1)
	default:
		meta = meta.SetRepSize(0)
	}
	switch {
	case posLen > math.MaxUint8:
		meta = meta.SetPosLenSize(1)
	default:
		meta = meta.SetPosLenSize(0)
	}
	switch {
	case bufLen > math.MaxUint32:
		meta = meta.SetBufLenSize(3)
	case bufLen > math.MaxUint16:
		meta = meta.SetBufLenSize(2)
	case bufLen > math.MaxUint8:
		meta = meta.SetBufLenSize(1)
	default:
		meta = meta.SetBufLenSize(0)
	}
	maxPos := s.MaxPos
	switch {
	case maxPos > math.MaxUint32:
		meta = meta.SetPosSize(3)
	case maxPos > math.MaxUint16:
		meta = meta.SetPosSize(2)
	case maxPos > math.MaxUint8:
		meta = meta.SetPosSize(1)
	default:
		meta = meta.SetPosSize(0)
	}
	// Start encoding.
	buffer = append(buffer, meta.ToByte())
	if s.Type == TypeRepeatSameChar {
		if s.Repeat > math.MaxUint8 {
			buffer = encoder.AppendUint16(buffer, uint16(s.Repeat))
		} else {
			buffer = append(buffer, byte(s.Repeat))
		}
	}
	switch meta.GetPosLenSize() {
	case 0:
		buffer = append(buffer, byte(posLen))
	case 1:
		buffer = encoder.AppendUint16(buffer, uint16(posLen))
	}
	for i := range s.Pos {
		switch meta.GetPosSize() {
		case 0:
			buffer = append(buffer, byte(s.Pos[i]))
		case 1:
			buffer = encoder.AppendUint16(buffer, uint16(s.Pos[i]))
		case 2:
			buffer = encoder.AppendUint32(buffer, uint32(s.Pos[i]))
		case 3:
			buffer = encoder.AppendUint64(buffer, uint64(s.Pos[i]))
		}
	}
	switch meta.GetBufLenSize() {
	case 0:
		buffer = append(buffer, byte(bufLen))
	case 1:
		buffer = encoder.AppendUint16(buffer, uint16(bufLen))
	case 2:
		buffer = encoder.AppendUint32(buffer, uint32(bufLen))
	case 3:
		buffer = encoder.AppendUint64(buffer, uint64(bufLen))
	}
	buffer = append(buffer, s.Buffer...)
	return buffer
}
