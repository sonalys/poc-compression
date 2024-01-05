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

// decompress2 is my attempt to make buffer decompression linear, so we can avoid storing position for segments with only 1 position.
// for this to work I need the iterator to fill the buffer from beginning to end, without utilizing the cur.pos[0].
func decompress(in block) []byte {
	out := make([]byte, 0, in.size)
	// first iteration, we fill all first positions to the out buffer without relying on the cur.pos variable.
	cur := in.head
	for {
		from := bytes.Repeat(cur.buffer, int(cur.repeat))
		cur.pos = cur.pos[1:]
		out = append(out, from...)
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	// from the second iteration next, we start using cur.pos to append data, however we need to right-shift the data.
	cur = in.head
	for {
		buf := bytes.Repeat(cur.buffer, int(cur.repeat))
		bufLen := uint32(len(buf))
		for _, pos := range cur.pos {
			// we only fill positions that are already addressable in the expanding buffer.
			if pos >= uint32(len(out)) {
				break
			}
			shift := pos + bufLen
			outLen := uint32(len(out))
			temp := make([]byte, outLen+bufLen)
			copy(temp, out)
			out = temp
			copy(out[shift:], out[pos:])
			copy(out[pos:], buf)
		}
		// fmt.Printf("%d %d\n", len(out), in.size)

		// loop until we fill the whole decompressed buffer
		if uint32(len(out)) == in.size {
			break
		} else if uint32(len(out)) > in.size {
			panic("memory leak")
		}
		if cur.next == nil {
			cur = in.head
		}
		cur = cur.next
	}
	return out
}

// TODO: maybe we can convert the uint32 coordinate system to a decompress order system,
// As long as we decompress in the right order, no need to store original pos.
