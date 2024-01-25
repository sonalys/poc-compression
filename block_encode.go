package gompressor

import (
	"encoding/binary"

	ll "github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
)

var encoder = binary.BigEndian

func Encode(b *Block) []byte {
	out := make([]byte, 0, 8+len(b.Buffer))
	out = encoder.AppendUint64(out, uint64(b.OriginalSize))
	out = encoder.AppendUint64(out, uint64(len(b.Buffer)))
	out = append(out, b.Buffer...)
	b.Segments.ForEach(func(cur *ll.ListEntry[segments.Segment]) {
		out = append(out, cur.Value.Encode()...)
	})
	return out
}
