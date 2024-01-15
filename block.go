package gompressor

import "golang.org/x/exp/constraints"

type BlockSize = constraints.Unsigned

type Block struct {
	Size   int64
	List   *LinkedList[Segment]
	Buffer []byte
}
