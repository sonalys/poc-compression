package segments

import "sort"

func getSegmentType(b byte) SegmentType {
	return SegmentType(b & 0b11)
}

func Byte2Bool(b byte) bool {
	return b != 0
}

func Bool2Byte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func MapBytePos(in []byte) (repetition [256][]int) {
	for pos, char := range in {
		if repetition[char] == nil {
			repetition[char] = make([]int, 0, 10)
		}
		repetition[char] = append(repetition[char], int(pos))
	}
	return
}

func GetBytePopularity(in [256][]int) []int {
	type bytePop struct {
		Char byte
		Len  int
	}
	var bytePopularity = make([]bytePop, 256)
	for char := 0; char < 256; char++ {
		bytePopularity[char] = bytePop{
			Char: byte(char),
			Len:  len(in[char]),
		}
	}
	sort.Slice(bytePopularity, func(i, j int) bool {
		return bytePopularity[i].Len > bytePopularity[j].Len
	})
	var result []int = make([]int, 256)
	for i := range bytePopularity {
		result[i] = int(bytePopularity[i].Char)
	}
	return result
}
