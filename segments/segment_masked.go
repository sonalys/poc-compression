package segments

import (
	"math"
)

type SegmentMasked struct {
	buffer       []byte
	pos          int
	byteCount    int
	mask         []mask
	enableInvert bool
}

func NewMaskedSegment(buffer []byte, pos int) *SegmentMasked {
	seg := &SegmentMasked{
		pos: pos,
	}
	var masks []*maskCalculator
	for j := 1; j < 4; j++ {
		masks = append(masks, newMaskCalculator(j))
	}
	for pos, b := range buffer {
		for j := 0; j < 3; j++ {
			masks[j].registerByte(b, len(buffer), pos)
		}
	}
	bestGain := math.MinInt64
	var bestMasks []mask
	for _, mask := range masks {
		if gain := mask.calcGain(); gain > bestGain {
			bestGain = gain
			bestMasks = mask.masks
		}
	}
	seg.mask = bestMasks
	// seg.mask, seg.enableInvert, seg.buffer = compression.CompressBuffer(buffer)
	seg.byteCount = len(buffer)
	return seg
}

func (s *SegmentMasked) Decompress(pos int) []byte {
	// return compression.DecompressBuffer(s.mask, s.enableInvert, s.buffer, s.byteCount)
	return nil
}

// func calculateMaskedCompressedSize(mask byte, enableInvert bool, uncompressedBufLen, pos int) int {
// 	compressedSize := 2
// 	compressedSize += getStorageByteSize(uncompressedBufLen)
// 	compressedSize += getStorageByteSize(pos)
// 	maskSize := compression.Count1Bits(mask)
// 	if enableInvert {
// 		maskSize++
// 	}
// 	compressedSize += (uncompressedBufLen*maskSize + 8 - 1) / 8
// 	return compressedSize
// }

func (s *SegmentMasked) getCompressedSize() int {
	// return calculateMaskedCompressedSize(s.mask, s.enableInvert, s.byteCount, s.pos)
	return 0
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
		InvertMask: s.enableInvert,
		PosSize:    NewMaxSize(s.pos),
		BufLenSize: NewMaxSize(s.byteCount),
	}
	buffer = append(buffer, meta.ToByte())
	// buffer = append(buffer, s.mask...)
	for i := range s.mask {
		buffer = append(buffer, s.mask[i].mask)
	}
	buffer = encodingFunc[meta.PosSize](buffer, s.pos)
	buffer = encodingFunc[meta.BufLenSize](buffer, s.byteCount)
	buffer = append(buffer, s.buffer...)
	return buffer
}

func DecodeMasked(b []byte) (*SegmentMasked, int) {
	var pos int
	meta, pos := NewMaskedGroupMeta(b[pos]), pos+1
	cur := SegmentMasked{
		enableInvert: meta.InvertMask,
	}
	// cur.mask, pos = b[pos], pos+1
	cur.pos, pos = decodeFunc[meta.PosSize](b, pos)
	cur.byteCount, pos = decodeFunc[meta.BufLenSize](b, pos)
	// maskSize := compression.Count1Bits(cur.mask)
	// if cur.enableInvert {
	// 	maskSize++
	// }
	// if maskSize == 8 {
	// 	cur.buffer = b[pos : pos+cur.byteCount]
	// 	return &cur, pos + cur.byteCount
	// }
	// if maskSize == 0 {
	// 	cur.buffer = make([]byte, 0)
	// 	return &cur, pos
	// }
	// compLen := (maskSize*cur.byteCount + 8 - 1) / 8
	// cur.buffer, pos = b[pos:pos+compLen], pos+compLen
	return &cur, pos
}
