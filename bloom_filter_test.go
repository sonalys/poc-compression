package gompressor

import (
	"os"
	"testing"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/stretchr/testify/require"
)

func getKeySizeBits(n int) int {
	if n >= 0b10000000 {
		return 8
	}
	if n >= 0b1000000 {
		return 7
	}
	if n >= 0b100000 {
		return 6
	}
	if n >= 0b10000 {
		return 5
	}
	if n >= 0b1000 {
		return 4
	}
	if n >= 0b100 {
		return 3
	}
	if n >= 0b10 {
		return 2
	}
	return 1
}

func Test_DictionaryRegeneration(t *testing.T) {
	t.Run("bloom test", func(t *testing.T) {
		const path string = "" +
			// "/bin/zsh"
			"/home/raicon/Downloads/snake.com"
		in, err := os.ReadFile(path)
		require.NoError(t, err)

		filter := bloom.New(64*3, 8)
		for _, char := range in {
			filter.Add([]byte{char})
		}
		maskSizeBits := len(filter.BitSet().Bytes()) * 8
		t.Logf("bloom filter size: %d", maskSizeBits)
		reg := RegenerateDict(filter)
		regLen := len(reg)
		t.Logf("dict[%d] = %v", regLen, reg)

		originalSize := len(in)
		compressedSize := (len(in)*getKeySizeBits(regLen) + maskSizeBits) / 8
		require.Less(t, compressedSize, originalSize)
	})
}
