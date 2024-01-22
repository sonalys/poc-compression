package segments

import "encoding/binary"

var encoder = binary.BigEndian

var encodingFunc = []func(buffer []byte, value int) []byte{
	func(buffer []byte, value int) []byte { return append(buffer, byte(value)) },
	func(buffer []byte, value int) []byte { return encoder.AppendUint16(buffer, uint16(value)) },
	func(buffer []byte, value int) []byte { return encoder.AppendUint32(buffer, uint32(value)) },
	func(buffer []byte, value int) []byte { return encoder.AppendUint64(buffer, uint64(value)) },
}

func encodePos(maxPos int, posList []int) []byte {
	posSize := NewMaxSize(maxPos)
	buffer := make([]byte, 0, int(posSize+1)*len(posList))
	for i := range posList {
		buffer = encodingFunc[posSize](buffer, posList[i])
	}
	return buffer
}
