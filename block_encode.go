package gompressor

import "encoding/binary"

var encoder = binary.BigEndian

func Encode(b *Block) []byte {
	out := make([]byte, 0, 8+len(b.Buffer))
	out = encoder.AppendUint64(out, uint64(b.OriginalSize))
	out = encoder.AppendUint64(out, uint64(len(b.Buffer)))
	out = append(out, b.Buffer...)
	if b.List == nil {
		return out
	}
	cur := b.List.Head
	for {
		if cur == nil {
			break
		}
		out = append(out, cur.Value.Encode()...)
		cur = cur.Next
	}
	return out
}
