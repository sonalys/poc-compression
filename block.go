package gompressor

import "golang.org/x/exp/constraints"

type BlockSize = constraints.Unsigned

type Block[S BlockSize] struct {
	Size   S
	List   *LinkedList[Segment[S]]
	Buffer []byte
}
