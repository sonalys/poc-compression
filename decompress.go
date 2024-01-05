package gompressor

import (
	"bytes"
)

// func decompress(in block) []byte {
// 	out := make([]byte, in.size)
// 	cur := in.head
// 	for {
// 		buf := bytes.Repeat(cur.buffer, int(cur.repeat))
// 		for _, pos := range cur.pos {
// 			copy(out[pos:], buf)
// 		}
// 		if cur.next == nil {
// 			break
// 		}
// 		cur = cur.next
// 	}
// 	return out
// }

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

// decompress2 is my attempt to make buffer decompression linear, so we can avoid storing position for segments with only 1 position.
// for this to work I need the iterator to fill the buffer from beginning to end, without utilizing the cur.pos[0].
func decompress(in *block) []byte {
	out := make([]byte, 0, in.size)
	for _, entry := range getOrderedDecompressionList(in.head) {
		out = append(out, bytes.Repeat(entry.buffer, int(entry.repeat))...)
	}
	return out
}

// TODO: maybe we can convert the uint32 coordinate system to a decompress order system,
// As long as we decompress in the right order, no need to store original pos.
