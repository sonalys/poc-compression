package gompressor

func Decompress(b *Block) []byte {
	out := make([]byte, b.OriginalSize)
	copy(out, b.Buffer)
	orderedSegments := sortAndFilterSegments(b.List, true)
	for _, cur := range orderedSegments {
		buf := cur.Decompress()
		lenBuf := int(len(buf))
		// right-shift data.
		copy(out[cur.Pos+lenBuf:], out[cur.Pos:])
		// copy out decompressed buf into out[pos].
		copy(out[cur.Pos:], buf)
	}
	return out
}
