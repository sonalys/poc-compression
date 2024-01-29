package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Compression(t *testing.T) {
	t.Run("same char", func(t *testing.T) {
		block := Compress([]byte{255, 255, 255, 255})
		got := Decompress(block)
		require.Equal(t, []byte{255, 255, 255, 255}, got)
	})
}
