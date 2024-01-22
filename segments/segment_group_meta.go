package segments

import "fmt"

type MetaRepeatGroup struct {
	Type       SegmentType // 2 bits
	InvertMask bool        // 1 bit
	PosLenSize MaxSize     // 1 bit
	BufLenSize MaxSize     // 2 bits
	PosSize    MaxSize     // 2 bits
}

func NewRepeatGroupMeta(b byte) MetaRepeatGroup {
	return MetaRepeatGroup{
		Type:       getSegmentType(b),
		InvertMask: Byte2Bool((b & (0b01 << 2) >> 2)),
		PosLenSize: MaxSize((b & (0b01 << 3) >> 3)),
		PosSize:    MaxSize((b & (0b11 << 4) >> 4)),
		BufLenSize: MaxSize((b & (0b11 << 5) >> 6)),
	}
}

func (m MetaRepeatGroup) Validate() error {
	if m.PosLenSize > 1 {
		return fmt.Errorf("posLen is bigger than 2 bytes")
	}
	return nil
}

func (m MetaRepeatGroup) ToByte() byte {
	if err := m.Validate(); err != nil {
		panic(err.Error())
	}
	var resp byte
	resp |= byte(m.Type)
	resp |= Bool2Byte(m.InvertMask) << 2
	resp |= byte(m.PosLenSize) << 3
	resp |= byte(m.PosSize) << 4
	resp |= byte(m.BufLenSize) << 6
	return resp
}
