package gompressor

type Block struct {
	Size   uint32
	List   *LinkedList[Segment]
	Buffer []byte
}
