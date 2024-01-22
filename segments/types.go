package segments

import (
	"math"
)

type (
	SegmentType byte
	MaxSize     byte
)

const (
	TypeMasked SegmentType = iota
	TypeGroup
	TypeSameChar

	MaxSizeUint8 MaxSize = iota
	MaxSizeUint16
	MaxSizeUint32
	MaxSizeUint64
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
