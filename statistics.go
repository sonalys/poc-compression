package gompressor

import "math"

func countBytes(input []byte) (repetition []uint) {
	repetition = make([]uint, 256)
	for _, char := range input {
		repetition[char] = repetition[char] + 1
	}
	return
}

func calculateSmallestDelta(input []byte, offset uint32, windowSize uint32) (uint32, float64) {
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
