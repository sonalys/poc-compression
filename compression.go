package gompressor

import (
	ll "github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
)

var layers = []func([]byte) (*ll.LinkedList[segments.Segment], []byte){
	segments.CreateSameCharSegments,
	segments.CreateGroupSegments,
	segments.CreateMaskedSegments,
}

func Compress(in []byte) *Block {
	list := ll.NewLinkedList[segments.Segment]()
	buffer := in
	for _, compressionLayer := range layers {
		newSegments, out := compressionLayer(buffer)
		list = newSegments.Append(list.Head)
		buffer = out
	}
	return &Block{
		OriginalSize: len(in),
		Segments:     list,
		Buffer:       buffer,
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
