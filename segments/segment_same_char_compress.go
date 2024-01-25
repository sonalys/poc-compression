package segments

import (
	"sort"

	ll "github.com/sonalys/gompressor/linkedlist"
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
	compressedSize := calculateSameCharCompressedSize(posLen, size, maxPos)
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

func CreateSameCharSegments(in []byte) []byte {
	byteMap := MapBytePos(in)
	list := &ll.LinkedList[*SegmentSameChar]{}
	const minSize = 2
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
			if size < minSize {
				continue
			}
			posBySize[size] = append(posBySize[size], posList[i])
			i = j - 1
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
			seg := NewRepeatSegment(size, byte(char), sp.positions[i]...)
			list.AppendValue(seg)
		}
	}
	raw := FillSegmentGaps(in, list)
	w := newBitWriter(make([]byte, 0, len(in)))
	cur := list.Head
	for {
		if cur == nil {
			break
		}
		w.Write(cur.Value.Encode())
		w.WriteByte(0b0, 1)
		cur = cur.Next
	}
	w.WriteByte(0b1, 1)

	orderedSegments := SortAndFilterSegments(list, false, removeBadSegments)
	var prev int
	for i, cur := range orderedSegments {
		if cur.Pos-prev > 0 {
			w.WriteByte(0b00, 2)
			w.Write(raw[prev:cur.Pos])
		}
		prev = cur.Pos
		w.WriteByte(0b01, 2)
		w.WriteByte(byte(i), 8)
	}
	if len(raw)-prev > 0 {
		w.WriteByte(0b00, 2)
		w.Write(raw[prev:])
	}
	w.WriteByte(0b11, 2)
	return w.buffer
}

type bitWriter struct {
	buffer []byte
	pos    int
}

func newBitWriter(buf []byte) bitWriter {
	return bitWriter{
		buffer: buf,
	}
}

func (b *bitWriter) WriteByte(in byte, size int) {
	bytePos := b.pos / 8
	if len(b.buffer) == bytePos {
		b.buffer = append(b.buffer, 0)
	}
	offset := b.pos + size - ((bytePos + 1) * 8)
	b.pos += size
	if offset <= 0 {
		b.buffer[bytePos] |= in << -offset
		return
	}
	b.buffer[bytePos] |= in >> offset
	b.buffer = append(b.buffer, in<<(8-offset))
}

func (b *bitWriter) ReadByte(pos, size int) byte {
	bytePos := pos / 8
	offset := pos + size - ((bytePos + 1) * 8)
	if offset <= 0 {
		value := b.buffer[bytePos] >> -offset
		return value
	}
	cleanOffset := 8 - size - offset
	value := b.buffer[bytePos]<<cleanOffset>>cleanOffset + b.buffer[bytePos+1]>>(8-offset)
	return value
}

func (b *bitWriter) Write(in []byte) {
	for _, value := range in {
		b.WriteByte(value, 8)
	}
}
