package gompressor

import "math"

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

func MapBytePos(in []byte) (repetition [256][]uint32) {
	for pos, char := range in {
		if repetition[char] == nil {
			repetition[char] = make([]uint32, 0, 100)
		}
		repetition[char] = append(repetition[char], uint32(pos))
	}
	return
}
