package gompressor

import (
	"fmt"
	"math"
)

type (
	SegmentType byte
	MaxSize     byte

	Meta struct {
		Type          SegmentType // 1 bit
		InvertBitmask bool        // 1 bit
		RepeatSize    MaxSize     // 1 bit
		PosLenSize    MaxSize     // 1 bit
		BufLenSize    MaxSize     // 2 bits
		PosSize       MaxSize     // 2 bits
	}

	// TODO: implement meta specific per segment type.
	// SameCharMetadata struct {
	// 	Type       SegmentType // 1 bit
	// 	RepeatSize MaxSize     // 1 bit
	// 	SinglePos  bool        // 1 bit
	// }

	// RepeatingGroupMetadata struct {
	// 	Type          SegmentType // 1 bit
	// 	BitMaskInvert bool        // 1 bit
	// }
)

const (
	MaxSizeUint8 MaxSize = iota
	MaxSizeUint16
	MaxSizeUint32
	MaxSizeUint64
)

const (
	TypeRepeatingGroup SegmentType = iota
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

func NewMeta2(b byte) Meta {
	return Meta{
		Type:          SegmentType(b & 0b01),
		InvertBitmask: Byte2Bool((b & (0b01 << 1) >> 1)),
		RepeatSize:    MaxSize((b & (0b01 << 2) >> 2)),
		PosLenSize:    MaxSize((b & (0b01 << 3) >> 3)),
		PosSize:       MaxSize((b & (0b11 << 4) >> 4)),
		BufLenSize:    MaxSize((b & (0b11 << 6) >> 6)),
	}
}

func (m Meta) Validate() error {
	if m.RepeatSize > 1 {
		return fmt.Errorf("repeat size is bigger than 2 bytes")
	}
	if m.PosLenSize > 1 {
		return fmt.Errorf("posLen is bigger than 2 bytes")
	}
	return nil
}

func (m Meta) ToByte() byte {
	if err := m.Validate(); err != nil {
		panic(err.Error())
	}
	var resp byte
	resp |= byte(m.Type)
	resp |= Bool2Byte(m.InvertBitmask) << 1
	resp |= byte(m.RepeatSize) << 2
	resp |= byte(m.PosLenSize) << 3
	resp |= byte(m.PosSize) << 4
	resp |= byte(m.BufLenSize) << 6
	return resp
}
