package bitbuffer

import "encoding/binary"

type BitBuffer struct {
	Buffer []byte
	pos    int
}

func NewBitBuffer(buf []byte) *BitBuffer {
	return &BitBuffer{
		Buffer: buf,
	}
}

func (b *BitBuffer) Write(in byte, byteSize int) byte {
	if byteSize < 0 {
		panic("byteSize cannot be negative")
	}
	bytePos := b.pos / 8
	if len(b.Buffer) == bytePos {
		b.Buffer = append(b.Buffer, 0)
	}
	offset := b.pos + byteSize - ((bytePos + 1) * 8)
	b.pos += byteSize

	value := in << (8 - byteSize) >> (8 - byteSize)
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

func (b *BitBuffer) Read(pos, byteSize int) byte {
	if byteSize > 8 {
		panic("size cannot be bigger than 8")
	}
	bytePos := pos / 8
	offset := pos + byteSize - ((bytePos + 1) * 8)
	if offset <= 0 {
		value := b.Buffer[bytePos] << (8 - byteSize + offset) >> (8 - byteSize)
		return value
	}
	value := b.Buffer[bytePos]<<(8-byteSize+offset)>>(8-byteSize) + b.Buffer[bytePos+1]>>(8-offset)
	return value
}

func (b *BitBuffer) WriteBuffer(in []byte) {
	for _, value := range in {
		b.Write(value, 8)
	}
}

func (b *BitBuffer) ReadBuffer(pos, len, byteSize int) []byte {
	out := make([]byte, 0, len)
	for i := 0; i < len; i++ {
		out = append(out, b.Read(pos, byteSize))
		pos += byteSize
	}
	return out
}

func GetBitUsage(value int) int {
	for i := 0; i < 64; i++ {
		if value <= 1<<i {
			return i
		}
	}
	panic("unreachable code")
}

func (b *BitBuffer) WriteCompact(in, size int) {
	r := NewBitBuffer(binary.BigEndian.AppendUint64(nil, uint64(in)))
	pos := 64 - size
	for i := pos; i <= 56; i += 8 {
		b.Write(r.Read(i, 8), 8)
		size -= 8
	}
	if size == 0 {
		return
	}
	b.Write(r.Read(64-size, size), size)
}

func (b *BitBuffer) ReadCompact(pos, size int) int {
	end := pos + size - 8
	buf := make([]byte, 8)
	bufStart := (64 - size) / 8
	for i := pos; i <= end; i += 8 {
		buf[bufStart] = b.Read(i, 8)
		size -= 8
		bufStart++
	}
	if size > 0 {
		buf[7] = b.Read(64-size, size)
	}
	return int(binary.BigEndian.Uint64(buf))
}
