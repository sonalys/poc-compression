package gompressor

import (
	ll "github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
	"golang.org/x/exp/constraints"
)

type BlockSize = constraints.Unsigned

type Block struct {
	OriginalSize int
	List         *ll.LinkedList[segments.Segment]
	Buffer       []byte
}
