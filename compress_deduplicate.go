package gompressor

import "bytes"

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
			if !bytes.Equal(curValue.Buffer, iterValue.Buffer) || curValue.Repeat != iterValue.Repeat || curValue.Type != iterValue.Type {
				goto end
			}
			// if pos doesn't overflow, we continue with the merge operation.
			if _, err := curValue.AddPos(iterValue.Pos); err == nil {
				iter.Remove()
			}
		end:
			iter = iter.Next
		}
		cur = cur.Next
	}
}
