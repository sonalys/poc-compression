package compression

import "bytes"

func createMask(in []byte) byte {
	var mask byte
	for _, char := range in {
		mask |= char
	}
	return mask
}

// Count1Bits counts how many 1 bits there are in a byte.
// Returns between 0 and 8
func Count1Bits(v byte) int {
	var count = 0
	for ; v != 0; count++ {
		v = v & (v - 1)
	}
	return count
}

func GetMaskBits(mask byte) []int {
	compressedBits := make([]int, 0, 8)
	for n := 0; n < 8; n++ {
		if mask&(1<<n) != 0 {
			compressedBits = append(compressedBits, n)
		}
	}
	return compressedBits
}

func CompressByte(compressBits []int, value byte) byte {
	var resp byte
	for i, n := range compressBits {
		resp |= ((value & (1 << n)) >> (n - i))
	}
	return resp
}

func DecompressByte(compressBits []int, value byte) byte {
	var resp byte
	for i, n := range compressBits {
		resp |= (value & (1 << i) << (n - i))
	}
	return resp
}

func CompressBuffer(in []byte) (byte, bool, []byte) {
	mask := createMask(in)
	maskSize := Count1Bits(mask)
	invert := false
	if maskSize == 0 {
		return mask, invert, []byte{}
	}
	if maskSize > 4 {
		buf := bytes.Clone(in)
		var invertMask byte
		for i := range buf {
			buf[i] = ^buf[i]
			invertMask |= buf[i]
		}
		if newSize := Count1Bits(invertMask); newSize < maskSize {
			if newSize == 0 {
				return invertMask, true, []byte{}
			}
			mask = invertMask
			maskSize = newSize
			invert = true
			in = buf
		}
	}
	compBitsSize := len(in) * maskSize
	compLen := (compBitsSize + 8 - 1) / 8
	compressed := make([]byte, compLen)
	compressedBits := GetMaskBits(mask)
	for i, char := range in {
		pos := i * maskSize
		bytePos := ((pos / 8) + 1) * 8
		offset := pos + maskSize - bytePos
		if offset <= 0 {
			compressed[pos/8] |= CompressByte(compressedBits, char) << -offset
			continue
		}
		compressed[pos/8] |= CompressByte(compressedBits, char) >> offset
		compressed[pos/8+1] |= CompressByte(compressedBits, char) << (8 - offset)
	}
	return mask, invert, compressed
}

func createFilterMask(maskSize int) byte {
	var filterMask byte
	for i := 0; i < maskSize; i++ {
		filterMask |= 1 << i
	}
	return filterMask
}

func DecompressBuffer(mask byte, invert bool, compressed []byte, compressedLen int) []byte {
	maskSize := Count1Bits(mask)
	if maskSize == 0 {
		buf := []byte{0}
		if invert {
			buf = []byte{0xff}
		}
		return bytes.Repeat(buf, compressedLen)
	}
	compressedBits := GetMaskBits(mask)
	buf := make([]byte, 0, compressedLen)
	filterMask := createFilterMask(maskSize)
	for pos := 0; pos < compressedLen*maskSize; pos += maskSize {
		bytePos := pos / 8
		offset := pos + maskSize - ((bytePos + 1) * 8)
		if offset <= 0 {
			buf = append(buf, DecompressByte(compressedBits, compressed[bytePos]&(filterMask<<-offset)>>-offset))
			continue
		}
		buf = append(buf, DecompressByte(compressedBits, compressed[bytePos]&(filterMask>>offset)<<offset+compressed[bytePos+1]>>(8-offset)))
	}
	if !invert {
		return buf
	}
	for i := range buf {
		buf[i] = ^buf[i]
	}
	return buf
}
