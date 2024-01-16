package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_encoding(t *testing.T) {
	in := []byte{255, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1, 255}
	expBlock := Compress(in)
	expSerialized := Encode(expBlock)
	gotBlock, err := Decode(expSerialized)
	require.NoError(t, err)
	require.Equal(t, expBlock, gotBlock)

	t.Run("reconstruction", func(t *testing.T) {
		got := Decompress(expBlock)
		require.Equal(t, in, got)
	})

	t.Run("deserialized reconstruction", func(t *testing.T) {
		out := Decompress(gotBlock)
		require.Equal(t, in, out)
	})

	t.Run("compression rate", func(t *testing.T) {
		ratio := float64(len(expSerialized)) / float64(len(in))
		if ratio > 1 {
			t.Errorf("compression increased file size. ratio: %.2f", ratio)
		}
	})
}
