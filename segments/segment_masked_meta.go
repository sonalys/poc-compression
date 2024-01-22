package segments

type MetaMaskedGroup struct {
	Type       SegmentType // 2 bits
	InvertMask bool        // 1 bit
	BufLenSize MaxSize     // 2 bits
	PosSize    MaxSize     // 2 bits
}

func NewMaskedGroupMeta(b byte) MetaMaskedGroup {
	return MetaMaskedGroup{
		Type:       getSegmentType(b),
		InvertMask: Byte2Bool((b & (0b01 << 2) >> 2)),
		PosSize:    MaxSize((b & (0b11 << 3) >> 3)),
		BufLenSize: MaxSize((b & (0b11 << 5) >> 5)),
	}
}

func (m MetaMaskedGroup) Validate() error {
	return nil
}

func (m MetaMaskedGroup) ToByte() byte {
	if err := m.Validate(); err != nil {
		panic(err.Error())
	}
	var resp byte
	resp |= byte(m.Type)
	resp |= Bool2Byte(m.InvertMask) << 2
	resp |= byte(m.PosSize) << 3
	resp |= byte(m.BufLenSize) << 5
	return resp
}
