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
				curValue.AppendPos(iterValue.Pos)
				iter.Remove()
			}
			iter = iter.Next
		}
		cur = cur.Next
	}
}
