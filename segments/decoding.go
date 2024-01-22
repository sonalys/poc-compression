package segments

import "encoding/binary"

var decoder = binary.BigEndian

var decodeFunc = []func(b []byte, pos int) (int, int){
	func(b []byte, pos int) (int, int) { return int(b[pos]), pos + 1 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint16(b[pos:])), pos + 2 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint32(b[pos:])), pos + 4 },
	func(b []byte, pos int) (int, int) { return int(decoder.Uint64(b[pos:])), pos + 8 },
}

func decodePos(b []byte, pos int, posLenSize, posSize MaxSize) (int, int, []int) {
	var posLen int16
	switch posLenSize {
	case MaxSizeUint8:
		posLen, pos = int16(b[pos]), pos+1
	case MaxSizeUint16:
		posLen, pos = int16(decoder.Uint16(b[pos:])), pos+2
	default:
		panic("invalid posSize")
	}
	var maxPos int
	posList := make([]int, posLen)
	posDecoder := decodeFunc[posSize]
	for i := range posList {
		posList[i], pos = posDecoder(b, pos)
		if posList[i] > maxPos {
			maxPos = posList[i]
		}
	}
	return maxPos, pos, posList
}
