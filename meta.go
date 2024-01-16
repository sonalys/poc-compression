package gompressor

type (
	Metadata    uint8
	SegmentType uint8
	Mask        uint8
)

const (
	TypeUncompressed SegmentType = iota
	TypeRepeatingGroup
	TypeRepeatSameChar

	SegmentTypeMask Mask = 0b11
	RepeatSizeMask  Mask = 0b1 << 2
	LenPosSizeMask  Mask = 0b1 << 3
	PosSizeMask     Mask = 0b11 << 4
	LenBufSizeMask  Mask = 0b11 << 6
)

func NewMetadata() Metadata {
	return Metadata(0)
}

func (m Metadata) Set(mask Mask, value byte) Metadata {
	return m&^Metadata(mask) | Metadata(byte(mask)&value)
}

func (m Metadata) Check(mask Mask) byte {
	return byte(m) & byte(mask)
}

func (m Metadata) ToByte() byte {
	return byte(m)
}

func (m Metadata) SetType(value SegmentType) Metadata {
	return m.Set(SegmentTypeMask, byte(value))
}

func (m Metadata) GetType() SegmentType {
	return SegmentType(m.Check(SegmentTypeMask))
}

func (m Metadata) SetRepSize(value byte) Metadata {
	return m.Set(RepeatSizeMask, byte(value)<<2)
}

func (m Metadata) SetPosLenSize(value byte) Metadata {
	return m.Set(LenPosSizeMask, byte(value)<<3)
}

func (m Metadata) SetPosSize(value byte) Metadata {
	return m.Set(PosSizeMask, byte(value)<<4)
}

func (m Metadata) SetBufLenSize(value byte) Metadata {
	return m.Set(LenBufSizeMask, byte(value)<<6)
}

func (m Metadata) GetRepSize() byte {
	return m.Check(RepeatSizeMask) >> 2
}
func (m Metadata) GetPosLenSize() byte {
	return m.Check(LenPosSizeMask) >> 3
}
func (m Metadata) GetPosSize() byte {
	return m.Check(PosSizeMask) >> 4
}
func (m Metadata) GetBufLenSize() byte {
	return m.Check(LenBufSizeMask) >> 6
}
