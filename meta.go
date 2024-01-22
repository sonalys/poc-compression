package gompressor

import (
	"fmt"
	"math"
)

type (
	SegmentType byte
	MaxSize     byte

	MetaSameChar struct {
		Type       SegmentType // 2 bits
		SinglePos  bool        // 1 bit
		RepeatSize MaxSize     // 1 bit
		PosLenSize MaxSize     // 1 bit
		PosSize    MaxSize     // 2 bits
	}

	MetaRepeatGroup struct {
		Type       SegmentType // 2 bits
		InvertMask bool        // 1 bit
		PosLenSize MaxSize     // 1 bit
		BufLenSize MaxSize     // 2 bits
		PosSize    MaxSize     // 2 bits
	}
)

const (
	MaxSizeUint8 MaxSize = iota
	MaxSizeUint16
	MaxSizeUint32
	MaxSizeUint64
)

const (
	TypeMasked SegmentType = iota
	TypeRepeatingGroup
	TypeRepeatSameChar
)

func NewMaxSize(value int) MaxSize {
	switch {
	case value > math.MaxUint32:
		return MaxSizeUint64
	case value > math.MaxUint16:
		return MaxSizeUint32
	case value > math.MaxUint8:
		return MaxSizeUint16
	default:
		return MaxSizeUint8
	}
}

func Byte2Bool(b byte) bool {
	return b != 0
}

func Bool2Byte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func getSegmentType(b byte) SegmentType {
	return SegmentType(b & 0b11)
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
