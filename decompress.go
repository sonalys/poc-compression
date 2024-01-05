package gompressor

import (
	"bytes"
)

func decompress(in block) []byte {
	out := make([]byte, in.size)
	cur := in.head
	for {
		from := bytes.Repeat(cur.buffer, int(cur.repeat))
		for _, pos := range cur.pos {
			copy(out[pos:], from)
		}
		if cur.next == nil {
			break
		}
		cur = cur.next
	}
	return out
}
