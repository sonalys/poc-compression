package gompressor

import "golang.org/x/exp/constraints"

type BlockSize = constraints.Unsigned

type Block struct {
	OriginalSize int
	List         *LinkedList[*Segment]
	Buffer       []byte
}
