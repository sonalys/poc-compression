package gompressor

// Deduplicate will find segments that are identical, besides position, and merge them.
func Deduplicate(list *LinkedList[Segment]) {
	cur := list.Head
	for {
		if cur == nil {
			break
		}
		iter := cur.Next
		for {
			if iter == nil {
				break
			}
			curValue := cur.Value
			iterValue := iter.Value
			if curValue.CanMerge(iterValue) {
				curSize := curValue.GetCompressedSize()
				iterSize := curValue.GetCompressedSize()
				// Merge into the node that is saving more space.
				if curSize < iterSize {
					curValue.AppendPos(iterValue.Pos)
					iter.Remove()
				} else {
					iterValue.AppendPos(curValue.Pos)
					cur.Remove()
					goto nextcur
				}
			}
			iter = iter.Next
		}
	nextcur:
		cur = cur.Next
	}
}
