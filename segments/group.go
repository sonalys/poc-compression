package segments

import (
	"github.com/sonalys/gompressor/compression"
)

type SegmentGroup struct {
	byteCount  int
	maxPos     int
	bitMask    byte
	invertMask bool
	buffer     []byte
	pos        []int
}

// NewGroupSegment creates a new segment.
func NewGroupSegment(buffer []byte, pos ...int) *SegmentGroup {
	seg := &SegmentGroup{
		buffer:    buffer,
		byteCount: len(buffer),
		bitMask:   0xff,
	}
	mask, invert, compressed := compression.CompressBuffer(buffer)
	if len(compressed) < len(buffer) {
		seg.bitMask = mask
		seg.invertMask = invert
		seg.buffer = compressed
	}
	return seg.appendPos(pos...)
}

func (s *SegmentGroup) appendPos(pos ...int) *SegmentGroup {
	for i := range pos {
		if pos[i] > s.maxPos {
			s.maxPos = pos[i]
		}
	}
	s.pos = append(s.pos, pos...)
	return s
}

func (s *SegmentGroup) Decompress() []byte {
	return compression.DecompressBuffer(s.bitMask, s.invertMask, s.buffer, s.byteCount)
}

func calculateGroupCompressedSize(posLen, bufLen, maxPos int) int {
	var compressedSize int = 2
	compressedSize += getStorageByteSize(posLen)
	compressedSize += getStorageByteSize(bufLen)
	compressedSize += posLen * getStorageByteSize(maxPos)
	compressedSize += bufLen
	return compressedSize
}

func (s *SegmentGroup) getCompressedSize() int {
	posLen := len(s.pos)
	bufLen := len(s.buffer)
	return calculateGroupCompressedSize(posLen, bufLen, s.maxPos)
}

func (s *SegmentGroup) GetOriginalSize() int {
	return s.byteCount * len(s.pos)
}

func (s *SegmentGroup) GetCompressionGains() int {
	return s.GetOriginalSize() - s.getCompressedSize()
}

func (s *SegmentGroup) GetPos() []int {
	return s.pos
}

func (s *SegmentGroup) GetType() SegmentType {
	return TypeGroup
}

func (s *SegmentGroup) Encode() []byte {
	buffer := make([]byte, 0, s.getCompressedSize())
	posLen := len(s.pos)
	meta := MetaRepeatGroup{
		Type:       TypeGroup,
		InvertMask: s.invertMask,
		PosLenSize: NewMaxSize(posLen),
		PosSize:    NewMaxSize(s.maxPos),
		BufLenSize: NewMaxSize(s.byteCount),
	}
	buffer = append(buffer, meta.ToByte())
	buffer = append(buffer, s.bitMask)
	buffer = encodingFunc[meta.PosLenSize](buffer, posLen)
	buffer = append(buffer, encodePos(s.maxPos, s.pos)...)
	buffer = encodingFunc[meta.BufLenSize](buffer, s.byteCount)
	buffer = append(buffer, s.buffer...)
	return buffer
}

func DecodeGroup(b []byte) (*SegmentGroup, int) {
	var pos int
	meta, pos := NewRepeatGroupMeta(b[pos]), pos+1
	cur := SegmentGroup{
		bitMask:    0xff,
		invertMask: meta.InvertMask,
	}
	cur.bitMask, pos = b[pos], pos+1
	cur.maxPos, pos, cur.pos = decodePos(b, pos, meta.PosLenSize, meta.PosSize)
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
