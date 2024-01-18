package gompressor

import (
	"math"
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

func MapBytePos(in []byte) (repetition [256][]int64) {
	for pos, char := range in {
		if repetition[char] == nil {
			repetition[char] = make([]int64, 0, 10)
		}
		repetition[char] = append(repetition[char], int64(pos))
	}
	return
}

func MapBytePosList(in []byte) (repetition [256]*LinkedList[int64]) {
	// t1 := time.Now()
	for pos, char := range in {
		if repetition[char] == nil {
			repetition[char] = &LinkedList[int64]{}
		}
		repetition[char].AppendValue(int64(pos))
	}
	// log.Debug().Str("duration", time.Since(t1).String()).Msg("byte pos indexing finished")
	return
}

func CalculateByteDensity(in []byte) (int, float64) {
	byteMap := make(map[byte]struct{}, 256)
	for _, char := range in {
		byteMap[char] = struct{}{}
	}
	return len(byteMap), float64(len(byteMap)) / float64(len(in))
}
