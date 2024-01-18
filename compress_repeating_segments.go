package gompressor

import (
	"bytes"
	"context"
	"sync"

	"golang.org/x/sync/semaphore"
)

// getEndOffset grows a byte group from the end position, finding the largest group possible.
// this group must repeat in curPos and nextPos.
func getEndOffset(buf []byte, curPos, nextPos int64) int64 {
	var endOffset int64 = 0
	bufLen := int64(len(buf))
	for newEnd := endOffset + 1; nextPos+newEnd < bufLen; newEnd++ {
		if buf[curPos+newEnd] == buf[nextPos+newEnd] {
			endOffset = newEnd
			continue
		}
		break
	}
	return endOffset
}

// matchRepeatingGroup finds all other positions from the same byte value that also repeat the group found.
func matchRepeatingGroup(cur *ListEntry[int64], buf, cmp []byte, endOffset int64) (resp []int64) {
	bufLen := int64(len(buf))
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	sem := semaphore.NewWeighted(10)
	for {
		if cur == nil {
			break
		}
		wg.Add(1)
		sem.Acquire(context.Background(), 1)
		go func(cur *ListEntry[int64]) {
			defer sem.Release(1)
			defer wg.Done()
			pos := cur.Value
			endPos := pos + endOffset
			startPos := pos
			// Prevent start and end positions from panicking,
			// also prevents position intersection from previous position with the current.
			if startPos > bufLen || endPos > bufLen {
				return
			}
			if bytes.Equal(buf[startPos:endPos], cmp) {
				lock.Lock()
				resp = append(resp, startPos)
				lock.Unlock()
			}
		}(cur)
		cur = cur.Next
	}
	wg.Wait()
	return
}

// func findBytePosIndex(list []int64, value int64) int {
// 	lower := 0
// 	higher := len(list) - 1
// 	for {
// 		if lower > higher {
// 			return -1
// 		}
// 		middle := lower + (higher-lower)/2
// 		// Check if x is present at mid
// 		if list[middle] == value {
// 			return middle
// 		} else if list[middle] < value {
// 			lower = middle + 1
// 		} else {
// 			higher = middle - 1
// 		}
// 	}
// }

// CreateRepeatingSegments should linearly detect repeating groups, without overlapping.
func CreateRepeatingSegments(buf []byte) (*LinkedList[*Segment], []byte) {
	// t1 := time.Now()
	bufLen := int64(len(buf))
	minSize := int64(3)
	list := NewLinkedList[*Segment]()
	byteMap := MapBytePosList(buf)
	collisionCheck := make([]*Segment, len(buf))
	for curPos := int64(0); curPos < bufLen; curPos++ {
		// percentage := (float64(curPos) / float64(bufLen)) * 100
		// log.Debug().Msgf("progress at %.2f%%", percentage)
		char := buf[curPos]
		bytePosList := byteMap[char]
		if seg := collisionCheck[curPos]; seg != nil {
			curPos += int64(len(seg.Buffer))*int64(seg.Repeat) - 1
			continue
		}
		posIndex := bytePosList.Find(curPos)
		next := posIndex
	nextBreak:
		for {
			next = next.Next
			if next == nil {
				break
			}
			nextPos := next.Value
			if seg := collisionCheck[nextPos]; seg != nil {
				for i := 0; i < int(len(seg.Buffer))*int(seg.Repeat)-1; i++ {
					next = next.Next
					if next == nil {
						break nextBreak
					}
				}
				continue
			}
			endOffset := getEndOffset(buf, curPos, nextPos)
			// Prevent positions too close to form a group.
			// Also prevents forming a group that intersects it's positions.
			if endOffset < minSize || endOffset > nextPos-curPos {
				continue
			}
			groupEnd := curPos + endOffset
			groupBuf := buf[curPos:groupEnd]
			// repeating group of minSize found in curPos and nextPos.
			matched := matchRepeatingGroup(next, buf, groupBuf, endOffset)
			matched = append(matched, curPos)
			nonCollidingPos := make([]int64, 0, len(matched))
			var bytesToRemove [256][]int64
		nextPos:
			for _, pos := range matched {
				// First check for byte collision with other segments.
				for k := range groupBuf {
					charPos := pos + int64(k)
					if collisionCheck[charPos] != nil {
						continue nextPos
					}
				}
				// No byte collisions means we can create our segment without any further issues.
				// So here register all these bytes to be removed from the indexing.
				for k, char := range groupBuf {
					charPos := pos + int64(k)
					bytesToRemove[char] = append(bytesToRemove[char], charPos)
				}
				// Position had no collisions, so it can be added to the segment.
				nonCollidingPos = append(nonCollidingPos, pos)
			}
			if len(nonCollidingPos) == 0 {
				continue
			}
			cur := NewSegment(TypeRepeatingGroup, groupBuf, nonCollidingPos...)
			// If the segment is net negative, it's not worth to add it.
			if cur.GetCompressionGains() <= 0 {
				continue
			}
			for char, posList := range bytesToRemove {
				for _, pos := range posList {
					// if posIndex == -1 {
					// 	panic("same byte mapped twice in the same segment")
					// }
					// byteMap[char] = append(byteMap[char][:posIndex], byteMap[char][posIndex+1:]...)
					byteMap[char].Find(pos).Remove()
					collisionCheck[pos] = cur
				}
			}
			// Adds the segment to the list and remove all the bytes used from indexing.
			list.AppendValue(cur)
			// We can skip this much position because it already contains a group.
			curPos = groupEnd
			break
		}
	}
	// log.Debug().Str("duration", time.Since(t1).String()).Int("segCount", list.Len).Msg("finishing repeatingSegments")
	return list, FillSegmentGaps(buf, list)
}
