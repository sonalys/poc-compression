package segments

import "fmt"

type MetaSameChar struct {
	Type       SegmentType // 2 bits
	SinglePos  bool        // 1 bit
	RepeatSize MaxSize     // 1 bit
	PosLenSize MaxSize     // 1 bit
	PosSize    MaxSize     // 2 bits
}

func NewSameCharMeta(b byte) MetaSameChar {
	return MetaSameChar{
		Type:       getSegmentType(b),
		SinglePos:  Byte2Bool((b & (0b01 << 2) >> 2)),
		RepeatSize: MaxSize((b & (0b01 << 3) >> 3)),
		PosLenSize: MaxSize((b & (0b01 << 4) >> 4)),
		PosSize:    MaxSize((b & (0b11 << 5) >> 5)),
	}
}

func (m MetaSameChar) Validate() error {
	if m.RepeatSize > 1 {
		return fmt.Errorf("repeat size is bigger than 2 bytes")
	}
	if m.PosLenSize > 1 {
		return fmt.Errorf("posLen is bigger than 2 bytes")
	}
	return nil
}

func (m MetaSameChar) ToByte() byte {
	if err := m.Validate(); err != nil {
		panic(err.Error())
	}
	var resp byte
	resp |= byte(m.Type)
	resp |= Bool2Byte(m.SinglePos) << 2
	resp |= byte(m.RepeatSize) << 3
	resp |= byte(m.PosLenSize) << 4
	resp |= byte(m.PosSize) << 5
	return resp
}
