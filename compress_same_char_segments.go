package gompressor

import (
	"math"
)

func CreateSameCharSegments(in []byte) (*LinkedList[*Segment], []byte) {
	// t1 := time.Now()
	lenIn := int64(len(in))
	list := &LinkedList[*Segment]{}
	// finds repetition groups and store them.
	for index := int64(0); index < lenIn; index++ {
		repeatCount := int64(1)
		for j := index + 1; j < lenIn && in[index] == in[j]; j++ {
			repeatCount += 1
			if repeatCount > math.MaxUint16 {
				panic("repeat overflow")
			}
		}
		if repeatCount < 2 {
			continue
		}
		list.AppendValue(NewRepeatSegment(index, uint16(repeatCount), []byte{in[index]}))
		index += repeatCount - 1
	}
	// log.Debug().Str("duration", time.Since(t1).String()).Int("segCount", list.Len).Msg("sameCharSegment")
	Deduplicate(list)
	return list, FillSegmentGaps(in, list)
}
