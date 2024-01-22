package gompressor

import (
	"math"
	"sort"

	"github.com/sonalys/gompressor/linkedlist"
)

func CountBytes(in []byte) (repetition []uint) {
	repetition = make([]uint, 256)
	for _, char := range in {
		repetition[char] = repetition[char] + 1
	}
	return
}

func CountRepetitions(in []byte) (repetition []map[int]uint) {
	repetition = make([]map[int]uint, 256)
	for i := range in {
		char := in[i]
		if repetition[char] == nil {
			repetition[char] = make(map[int]uint)
		}
		var size int = 1
		for j := i + 1; j < len(in) && in[j] == char; j++ {
			size++
		}
		repetition[char][size] += 1
	}
	return
}

func CalculateSmallestDelta(input []byte, offset uint32, windowSize uint32) (uint32, float64) {
	inputLen := uint32(len(input))

	var resp float64 = math.MaxFloat64
	var pos uint32
	// delta := (float64(inputLen) / float64(windowSize))
	for i := offset; i+windowSize < inputLen; i += 1 {
		start := i
		end := i + windowSize
		chunk := input[start:end]

		var sum float64
		for j := 1; j < len(chunk); j++ {
			sum += float64(chunk[j] - chunk[j-1])
		}
		avg := sum / float64(windowSize)
		if avg < resp {
			resp = avg
			pos = uint32(i)
		}
	}
	return pos, resp
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

func MapBytePosList(in []byte) (repetition [256]*linkedlist.LinkedList[int]) {
	// t1 := time.Now()
	for pos, char := range in {
		if repetition[char] == nil {
			repetition[char] = &linkedlist.LinkedList[int]{}
		}
		repetition[char].AppendValue(int(pos))
	}
	// log.Debug().Str("duration", time.Since(t1).String()).Msg("byte pos indexing finished")
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

func CalculateByteDensity(in []byte) (int, float64) {
	byteMap := make(map[byte]struct{}, 256)
	for _, char := range in {
		byteMap[char] = struct{}{}
	}
	return len(byteMap), float64(len(byteMap)) / float64(len(in))
}
