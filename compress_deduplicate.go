package gompressor

// Deduplicate will find segments that are identical, besides position, and merge them.
func Deduplicate(list *LinkedList[*Segment]) int {
	// t1 := time.Now()
	var count int
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
				count++
				// Merge into the node that is saving more space.
				if curSize < iterSize {
					curValue.AppendPos(iterValue.Pos...)
					iter.Remove()
				} else {
					iterValue.AppendPos(curValue.Pos...)
					cur.Remove()
					goto nextcur
				}
			}
			iter = iter.Next
		}
	nextcur:
		cur = cur.Next
	}
	// log.Debug().
	// 	Str("duration", time.Since(t1).String()).
	// 	int("deduplicatedCount", count).
	// 	int("segCount", int(list.Len)).
	// 	Msg("deduplicate")
	return count
}
