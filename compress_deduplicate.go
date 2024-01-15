package gompressor

// Deduplicate will find segments that are identical, besides position, and merge them.
func Deduplicate[S BlockSize](list *LinkedList[Segment[S]]) {
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
				if _, err := curValue.AppendPos(iterValue.Pos); err == nil {
					iter.Remove()
				}
			}
			iter = iter.Next
		}
		cur = cur.Next
	}
}
