package gompressor

func countBytes(input []byte) (repetition []uint) {
	repetition = make([]uint, 256)
	for _, char := range input {
		repetition[char] = repetition[char] + 1
	}
	return
}
