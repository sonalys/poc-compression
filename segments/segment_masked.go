package segments

import (
	"github.com/sonalys/gompressor/bitbuffer"
	"github.com/sonalys/gompressor/compression"
)

type SegmentMasked struct {
	buffer       []byte
	pos          int
	byteCount    int
	bitMask      byte
	enableInvert bool
}

func NewMaskedSegment(buffer []byte, pos int) *SegmentMasked {
	seg := &SegmentMasked{
		pos: pos,
	}
	seg.bitMask, seg.enableInvert, seg.buffer = compression.CompressBuffer(buffer)
	seg.byteCount = len(buffer)
	return seg
}

func (s *SegmentMasked) Decompress(pos int) []byte {
	return compression.DecompressBuffer(s.bitMask, s.enableInvert, s.buffer, s.byteCount)
}

func calculateMaskedCompressedSize(mask byte, enableInvert bool, uncompressedBufLen, pos int) int {
	compressedSize := 2
	maskSize := compression.Count1Bits(mask)
	if enableInvert {
		maskSize++
	}
	compressedSize += (uncompressedBufLen*maskSize + 8 - 1) / 8
	return compressedSize
}

func (s *SegmentMasked) getCompressedSize() int {
	return calculateMaskedCompressedSize(s.bitMask, s.enableInvert, s.byteCount, s.pos)
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

func (s *SegmentMasked) Encode(w *bitbuffer.BitBuffer) {
	w.Write(s.bitMask, 8)
	w.WriteBuffer(encodingFunc[0](nil, s.byteCount))
	w.WriteBuffer(s.buffer)
}

func DecodeMasked(b []byte) (*SegmentMasked, int) {
	var pos int
	meta, pos := NewMaskedGroupMeta(b[pos]), pos+1
	cur := SegmentMasked{
		enableInvert: meta.InvertMask,
	}
	cur.bitMask, pos = b[pos], pos+1
	cur.pos, pos = decodeFunc[meta.PosSize](b, pos)
	cur.byteCount, pos = decodeFunc[meta.BufLenSize](b, pos)
	maskSize := compression.Count1Bits(cur.bitMask)
	if cur.enableInvert {
		maskSize++
	}
	if maskSize == 8 {
		cur.buffer = b[pos : pos+cur.byteCount]
		return &cur, pos + cur.byteCount
	}
	if maskSize == 0 {
		cur.buffer = make([]byte, 0)
		return &cur, pos
	}
	compLen := (maskSize*cur.byteCount + 8 - 1) / 8
	cur.buffer, pos = b[pos:pos+compLen], pos+compLen
	return &cur, pos
}
