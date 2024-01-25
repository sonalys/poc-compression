package bitbuffer

type BitBuffer struct {
	Buffer []byte
	pos    int
}

func NewBitBuffer(buf []byte) BitBuffer {
	return BitBuffer{
		Buffer: buf,
	}
}

func (b *BitBuffer) Write(in byte, size int) byte {
	bytePos := b.pos / 8
	if len(b.Buffer) == bytePos {
		b.Buffer = append(b.Buffer, 0)
	}
	offset := b.pos + size - ((bytePos + 1) * 8)
	b.pos += size

	value := in << (8 - size) >> (8 - size)
	if value != in {
		panic("in is overflowing given size")
	}
	if offset <= 0 {
		b.Buffer[bytePos] |= in << -offset
		return in
	}
	b.Buffer[bytePos] |= in >> offset
	b.Buffer = append(b.Buffer, in<<(8-offset))
	return in
}

func (b *BitBuffer) Read(pos, size int) byte {
	if size > 8 {
		panic("size cannot be bigger than 8")
	}
	bytePos := pos / 8
	offset := pos + size - ((bytePos + 1) * 8)
	if offset <= 0 {
		value := b.Buffer[bytePos] << (8 - size + offset) >> (8 - size)
		return value
	}
	value := b.Buffer[bytePos]<<(8-size+offset)>>(8-size) + b.Buffer[bytePos+1]>>(8-offset)
	return value
}

func (b *BitBuffer) WriteBuffer(in []byte) {
	for _, value := range in {
		b.Write(value, 8)
	}
}
