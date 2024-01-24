package compression

import "bytes"

func MaskRegisterBuffer(buffer []byte) (mask byte, maskSize int, enableInvert bool, invertList []bool) {
	invertList = make([]bool, len(buffer))
	for i, b := range buffer {
		var shouldEnableInvert bool
		var byteInvert bool
		mask, shouldEnableInvert, byteInvert, maskSize = MaskRegisterByte(mask, b)
		enableInvert = enableInvert || shouldEnableInvert
		if maskSize == 8 || shouldEnableInvert && maskSize == 7 {
			return
		}
		if byteInvert {
			invertList[i] = true
		}
	}
	return
}

func MaskRegisterByte(m byte, value byte) (mask byte, enableInvert, byteInvert bool, maskSize int) {
	normal := m | value
	inverted := m | ^value
	// If we are using 7 bits or more on both masks, we won't save any space.
	// So we just return the original input with a full mask.
	if normal == 255 && inverted == 255 {
		return 0xff, false, false, 8
	}
	invertedSize := Count1Bits(inverted)
	normalSize := Count1Bits(normal)
	if invertedSize < normalSize {
		return inverted, true, true, invertedSize
	}
	return normal, false, false, normalSize
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

func Byte2Bool(b byte) bool {
	return b != 0
}

func Bool2Byte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func CompressByte(compressBits []int, enableInvert, shouldInvertByte bool, value byte) (resp byte) {
	if shouldInvertByte {
		value = ^value
	}
	for i, n := range compressBits {
		resp |= value & (1 << n) >> (n - i)
	}
	if enableInvert {
		return resp<<1 | Bool2Byte(shouldInvertByte)
	}
	return resp
}

func DecompressByte(compressBits []int, enableInvert bool, value byte) (resp byte) {
	var didByteInvert byte
	if enableInvert {
		didByteInvert = value & 0b1
		value = value >> 1
	}
	for i, n := range compressBits {
		resp |= value & (1 << i) << (n - i)
	}
	if didByteInvert != 0 {
		return ^resp
	}
	return resp
}

func CompressBuffer(in []byte) (mask byte, enableInvert bool, compressed []byte) {
	mask, maskSize, enableInvert, invertList := MaskRegisterBuffer(in)
	if maskSize == 8 || enableInvert && maskSize == 7 {
		return 0xff, enableInvert, in
	}
	if !enableInvert && maskSize == 0 {
		return 0x00, false, []byte{}
	}
	// maskSize + 1 because we use bit 0 for invert flag.
	if enableInvert {
		maskSize++
	}
	compBitsSize := len(in) * maskSize
	compLen := (compBitsSize + 8 - 1) / 8
	compressed = make([]byte, compLen)
	compressedBits := GetMaskBits(mask)
	for i, char := range in {
		pos := i * maskSize
		bytePos := ((pos / 8) + 1) * 8
		offset := pos + maskSize - bytePos
		if offset <= 0 {
			compressed[pos/8] |= CompressByte(compressedBits, enableInvert, invertList[i], char) << -offset
			continue
		}
		value := CompressByte(compressedBits, enableInvert, invertList[i], char)
		compressed[pos/8] |= value >> offset
		compressed[pos/8+1] |= value << (8 - offset)
	}
	return mask, enableInvert, compressed
}

func createFilterMask(maskSize int) byte {
	var filterMask byte
	for i := 0; i < maskSize; i++ {
		filterMask |= 1 << i
	}
	return filterMask
}

func DecompressBuffer(mask byte, enableInvert bool, compressed []byte, compressedLen int) []byte {
	if mask == 0xff {
		return compressed
	}
	maskSize := Count1Bits(mask)
	if !enableInvert && maskSize == 0 {
		buf := []byte{0}
		return bytes.Repeat(buf, compressedLen)
	}
	if enableInvert {
		maskSize++
	}
	compressedBits := GetMaskBits(mask)
	buf := make([]byte, 0, compressedLen)
	filterMask := createFilterMask(maskSize)
	for pos := 0; pos < compressedLen*maskSize; pos += maskSize {
		bytePos := pos / 8
		offset := pos + maskSize - ((bytePos + 1) * 8)
		if offset <= 0 {
			value := compressed[bytePos] & (filterMask << -offset) >> -offset
			buf = append(buf, DecompressByte(compressedBits, enableInvert, value))
			continue
		}
		value := compressed[bytePos]&(filterMask>>offset)<<offset + compressed[bytePos+1]>>(8-offset)
		buf = append(buf, DecompressByte(compressedBits, enableInvert, value))
	}
	return buf
}
