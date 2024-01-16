package gompressor

// Metadata is a bitmask used to store metadata.
// Address	|		data
//
//	1, 2					Segment types = max 4.
//	3*						If type Repeat, 0 = 1 byte, 1 = 2 bytes for seg.Repeat.
//	4							lenPos size, 0 = 1 byte, 1 = 2 bytes
//	5, 6					lenBuf size, 0 = 1 byte, 1 = 2 bytes, 2 = 4 bytes, 3 = 8 bytes
//	7, 8					pos    size, 0 = 1 byte, 1 = 2 bytes, 3 = 4 bytes, 4 = 8 bytes
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
	LenPosSizeMask  Mask = 0b1 << 4
	PosSizeMask     Mask = 0b11 << 5
	LenBufSizeMask  Mask = 0b11 << 6
)

func NewMetadata() *Metadata {
	m := Metadata(0)
	return &m
}

func (m *Metadata) Set(mask Mask, value byte) *Metadata {
	// Clear bits and then set value.
	*m = Metadata(byte(*m)&^byte(mask) | byte(mask)&value)
	return m
}

func (m *Metadata) Check(mask Mask) byte {
	return byte(*m) & byte(mask)
}

func (m *Metadata) ToByte() byte {
	return byte(*m)
}
