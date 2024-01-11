package gompressor

// meta is a bitmask used to store metadata.
// Address	|		data
//
//	1, 2					Segment types = max 4.
//	3							Repeat size, 0 = 1 byte, 1 = 2 bytes.
//	4,5,6,7,8			posLen = max 32.
type meta uint8
type SegmentType uint8

const (
	TypeUncompressed SegmentType = iota
	TypeRepeatingGroup
	TypeRepeatSameChar

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
func (m meta) setType(t SegmentType) meta { return (m & 0b11111100) | meta(t) }

// getType
// clear all bytes except 2 and 3
// shift right 1 byte to get segType
func (m meta) getType() SegmentType { return SegmentType(m & 0b11) }

func (m meta) isRepeat2Bytes() bool { return m&flagRepeatIs2Bytes != 0 }
func (m meta) setIsRepeat2Bytes(value bool) meta {
	if value {
		return m | flagRepeatIs2Bytes
	}
	return m &^ flagRepeatIs2Bytes
}
