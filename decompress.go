package gompressor

import "github.com/sonalys/gompressor/segments"

func Decompress(b *Block) []byte {
	out := make([]byte, b.OriginalSize)
	copy(out, b.Buffer)
	orderedSegments := segments.SortAndFilterSegments(b.Segments, true)
	for _, cur := range orderedSegments {
		buffer := cur.Decompress()
		bufLen := len(buffer)
		// right-shift data.
		copy(out[cur.Pos+bufLen:], out[cur.Pos:])
		// copy out decompressed buf into out[pos].
		copy(out[cur.Pos:], buffer)
	}
	return out
}
