package gompressor

import (
	"bytes"
)

func getOrderedDecompressionList(list []orderedSegment) (out []orderedSegment) {
	out = make([]orderedSegment, 0, len(list))
	var order uint8 = 0
	for {
		found := false
		for _, entry := range list {
			for _, curOrder := range entry.order {
				if curOrder == order {
					out = append(out, entry)
					found = true
					order++
				}
			}
		}
		if !found {
			return
		}
	}
}

func Decompress(in *block) []byte {
	out := make([]byte, 0, in.size)
	for _, entry := range getOrderedDecompressionList(in.head) {
		out = append(out, bytes.Repeat(entry.buffer, int(entry.repeat))...)
	}
	return out
}
