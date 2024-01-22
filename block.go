package gompressor

import (
	"github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
	"golang.org/x/exp/constraints"
)

type BlockSize = constraints.Unsigned

type Block struct {
	OriginalSize int
	List         *linkedlist.LinkedList[segments.Segment]
	Buffer       []byte
}
