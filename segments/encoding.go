package segments

import (
	"encoding/binary"

	"github.com/sonalys/gompressor/bitbuffer"
	ll "github.com/sonalys/gompressor/linkedlist"
)

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

func encodeSegments[T Segment](inLen int, list *ll.LinkedList[T], raw []byte) []byte {
	w := bitbuffer.NewBitBuffer(make([]byte, 0, inLen))
	list.ForEach(func(cur *ll.ListEntry[T]) {
		cur.Value.Encode(w)
		w.Write(0b0, 1)
	})
	w.Write(0b1, 1)
	w.WriteBuffer(raw)
	return w.Buffer
}
