package segments

import "github.com/sonalys/gompressor/compression"

type SegmentMasked struct {
	buffer     []byte
	pos        int
	byteCount  int
	bitMask    byte
	invertMask bool
}

type MaskedSegmentOption interface {
	Apply(*SegmentMasked)
}

type WithBuffer []byte

type WithMask struct {
	ByteCount  int
	BitMask    byte
	InvertMask bool
	Compressed []byte
}

func (w WithBuffer) Apply(seg *SegmentMasked) {
	seg.bitMask, seg.invertMask, seg.buffer = compression.CompressBuffer(w)
	seg.byteCount = len(w)
}

func (w WithMask) Apply(seg *SegmentMasked) {
	seg.bitMask, seg.invertMask, seg.buffer = w.BitMask, w.InvertMask, w.Compressed
}

func NewMaskedSegment(opt MaskedSegmentOption, pos int) *SegmentMasked {
	seg := &SegmentMasked{
		pos: pos,
	}
	opt.Apply(seg)
	return seg
}

func (s *SegmentMasked) Decompress() []byte {
	return compression.DecompressBuffer(s.bitMask, s.invertMask, s.buffer, s.byteCount)
}

func calculateMaskedCompressedSize(mask byte, uncompressedBufLen, pos int) int {
	compressedSize := 2
	compressedSize += getStorageByteSize(uncompressedBufLen)
	compressedSize += getStorageByteSize(pos)
	compressedSize += (uncompressedBufLen*compression.Count1Bits(mask) + 8 - 1) / 8
	return compressedSize
}

func (s *SegmentMasked) getCompressedSize() int {
	return calculateMaskedCompressedSize(s.bitMask, s.byteCount, s.pos)
}

func (s *SegmentMasked) GetOriginalSize() int {
	return s.byteCount
}

func (s *SegmentMasked) GetCompressionGains() int {
	return s.GetOriginalSize() - s.getCompressedSize()
}

func (s *SegmentMasked) GetPos() []int {
	return []int{s.pos}
}

func (s *SegmentMasked) GetType() SegmentType {
	return TypeMasked
}

func (s *SegmentMasked) Encode() []byte {
	buffer := make([]byte, 0, s.getCompressedSize())
	meta := MetaMaskedGroup{
		Type:       TypeMasked,
		InvertMask: s.invertMask,
		PosSize:    NewMaxSize(s.pos),
		BufLenSize: NewMaxSize(s.byteCount),
	}
	buffer = append(buffer, meta.ToByte())
	buffer = append(buffer, s.bitMask)
	buffer = encodingFunc[meta.PosSize](buffer, s.pos)
	buffer = encodingFunc[meta.BufLenSize](buffer, s.byteCount)
	buffer = append(buffer, s.buffer...)
	return buffer
}

func DecodeMasked(b []byte) (*SegmentMasked, int) {
	var pos int
	meta, pos := NewMaskedGroupMeta(b[pos]), pos+1
	cur := SegmentMasked{
		bitMask:    0xff,
		invertMask: meta.InvertMask,
	}
	cur.bitMask, pos = b[pos], pos+1
	cur.pos, pos = decodeFunc[meta.PosSize](b, pos)
	cur.byteCount, pos = decodeFunc[meta.BufLenSize](b, pos)
	maskSize := compression.Count1Bits(cur.bitMask)
	if maskSize == 8 || maskSize == 0 {
		cur.buffer = b[pos : pos+cur.byteCount]
		return &cur, pos + cur.byteCount
	}
	compLen := (maskSize*cur.byteCount + 8 - 1) / 8
	cur.buffer, pos = b[pos:pos+compLen], pos+compLen
	return &cur, pos
}
