package gompressor

import (
	"bytes"
	"fmt"

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
	for i, compressionLayer := range layers {
		newSegments, out := compressionLayer(buffer)
		testBlock := &Block{
			OriginalSize: len(buffer),
			Segments:     newSegments,
			Buffer:       out,
		}
		// Check if decompressing current layer will revert the buffer back to the previous layer.
		if !bytes.Equal(Decompress(testBlock), buffer) {
			panic(fmt.Sprintf("fuck on layer %d", i))
		}
		list = newSegments.Append(list.Head)
		buffer = out
	}
	// Don't run any optimizations outside, because the coordinates of each layer are relative.
	// So if you merge 2 segments from different layers, in reality they have different coordinates.
	b := &Block{
		OriginalSize: len(in),
		Segments:     list,
		Buffer:       buffer,
	}
	return b
}
