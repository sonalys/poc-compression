package gompressor

import (
	"github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
)

func Compress(buf []byte) *Block {
	size := len(buf)
	layers := []func([]byte) (*linkedlist.LinkedList[segments.Segment], []byte){
		segments.CreateSameCharSegments,
		segments.CreateGroupSegments,
	}
	list := linkedlist.NewLinkedList[segments.Segment]()
	for _, compressionLayer := range layers {
		var newSegments *linkedlist.LinkedList[segments.Segment]
		// note that we are changing buf through each layer.
		// that means different coordinates for each layer.
		newSegments, buf = compressionLayer(buf)
		list.Append(newSegments.Head)
	}
	// Don't run any optimizations outside, because the coordinates of each layer are relative.
	// So if you merge 2 segments from different layers, in reality they have different coordinates.
	return &Block{
		OriginalSize: size,
		List:         list,
		Buffer:       buf,
	}
}
