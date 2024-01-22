package gompressor

import (
	"sort"
)

type sizePos struct {
	sizes     []int
	positions [][]int
}

func (sp *sizePos) Len() int {
	return len(sp.sizes)
}

func (sp *sizePos) Less(i int, j int) bool {
	return sp.sizes[i] < sp.sizes[j]
}

func (sp *sizePos) Swap(i int, j int) {
	sp.sizes[i], sp.sizes[j] = sp.sizes[j], sp.sizes[i]
	sp.positions[i], sp.positions[j] = sp.positions[j], sp.positions[i]
}

func (sp *sizePos) getPrevious(i int) int {
	var other int
	for j := i - 1; j >= 0; j-- {
		if len(sp.positions[j]) > 0 {
			other = j
			break
		}
	}
	return other
}

func getRepeatGain(i, posLen, size, maxPos int) int {
	originalSize := size * posLen
	compressedSize := GetCompressedSize(TypeRepeatSameChar, size, maxPos, posLen, 1)
	return originalSize - compressedSize
}

func shouldMerge(sp *sizePos, cur, other int) bool {
	curLenPos := len(sp.positions[cur])
	otherLenPos := len(sp.positions[other])
	curGain := getRepeatGain(cur, curLenPos, sp.sizes[cur], sp.positions[cur][curLenPos-1])
	otherGain := getRepeatGain(other, otherLenPos, sp.sizes[other], sp.positions[other][otherLenPos-1])
	maxPos := sp.positions[cur][curLenPos-1]
	if otherMaxPos := sp.positions[other][otherLenPos-1]; otherMaxPos > maxPos {
		maxPos = otherMaxPos
	}
	mergeGain := getRepeatGain(other, len(sp.positions[other])+curLenPos, sp.sizes[other], maxPos)
	return curGain+otherGain < mergeGain
}

func CreateSameCharSegments(buf []byte) (*LinkedList[*Segment], []byte) {
	byteMap := MapBytePos(buf)
	list := &LinkedList[*Segment]{}
	for char, posList := range byteMap {
		posBySize := make(map[int][]int, len(posList))
		for i := 0; i < len(posList); i++ {
			var j int
			for j = i + 1; j < len(posList); j++ {
				if posList[j] != posList[j-1]+1 {
					break
				}
			}
			size := j - i
			if size < 2 {
				continue
			}
			posBySize[size] = append(posBySize[size], posList[i])
			i += size - 1
		}
		sp := &sizePos{
			sizes:     make([]int, 0, len(posBySize)),
			positions: make([][]int, 0, len(posBySize)),
		}
		for size, posList := range posBySize {
			sp.sizes = append(sp.sizes, size)
			sp.positions = append(sp.positions, posList)
		}
		sort.Sort(sp)
		for i := len(sp.sizes) - 1; i > 0; i-- {
			if len(sp.positions[i]) == 0 {
				continue
			}
			other := sp.getPrevious(i)
			if shouldMerge(sp, i, other) {
				sp.positions[other] = append(sp.positions[other], sp.positions[i]...)
				sp.positions[i] = nil
			}
		}
		for i, size := range sp.sizes {
			if len(sp.positions[i]) == 0 {
				continue
			}
			seg := NewRepeatSegment(size, []byte{byte(char)}, sp.positions[i]...)
			list.AppendValue(seg)
		}
	}
	return list, FillSegmentGaps(buf, list)
}
