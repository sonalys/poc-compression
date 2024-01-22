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
	size := len(in)
	list := ll.NewLinkedList[segments.Segment]()
	for _, compressionLayer := range layers {
		var newSegments *ll.LinkedList[segments.Segment]
		newSegments, in = compressionLayer(in)
		list.Append(newSegments.Head)
	}
	// Don't run any optimizations outside, because the coordinates of each layer are relative.
	// So if you merge 2 segments from different layers, in reality they have different coordinates.
	return &Block{
		OriginalSize: size,
		List:         list,
		Buffer:       in,
	}
}
