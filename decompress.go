package gompressor

import "github.com/sonalys/gompressor/segments"

func Decompress(b *Block) []byte {
	out := make([]byte, b.OriginalSize)
	copy(out, b.Buffer)
	orderedSegments := segments.SortAndFilterSegments(b.List, true)
	for _, cur := range orderedSegments {
		buf := cur.Decompress()
		bufLen := len(buf)
		// right-shift data.
		copy(out[cur.Pos+bufLen:], out[cur.Pos:])
		// copy out decompressed buf into out[pos].
		copy(out[cur.Pos:], buf)
	}
	return out
}
