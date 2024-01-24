package gompressor

import (
	"github.com/sonalys/gompressor/segments"
)

var layers = []func([]byte) []byte{
	segments.CreateSameCharSegments,
	// segments.CreateGroupSegments,
	// segments.CreateMaskedSegments,
}

func Compress(in []byte) *Block {
	for _, compressionLayer := range layers {
		in = compressionLayer(in)
	}
	return &Block{
		OriginalSize: len(in),
		Buffer:       in,
	}
}

func Decompress(b *Block) []byte {
	out := make([]byte, b.OriginalSize)
	copy(out, b.Buffer)
	orderedSegments := segments.SortAndFilterSegments(b.Segments, true)
	for _, cur := range orderedSegments {
		buffer := cur.Decompress(cur.Pos)
		bufLen := len(buffer)
		// right-shift data.
		copy(out[cur.Pos+bufLen:], out[cur.Pos:])
		// copy out decompressed buf into out[pos].
		copy(out[cur.Pos:], buffer)
	}
	return out
}
