package segments

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetCompressionGain(t *testing.T) {
	t.Run("positive gain", func(t *testing.T) {
		seg := NewRepeatSegment(50, 1, 0, 100)
		originalSize := int(100)
		compressedSize := int(len(seg.Encode()))
		require.Equal(t, compressedSize, seg.getCompressedSize())
		gains := seg.GetCompressionGains()
		require.Equal(t, originalSize-compressedSize, gains)
	})

	t.Run("negative gain", func(t *testing.T) {
		seg := NewGroupSegment([]byte{1, 2}, 0, 100)
		originalSize := int(4)
		compressedSize := int(len(seg.Encode()))
		require.Equal(t, compressedSize, seg.getCompressedSize())
		gains := seg.GetCompressionGains()
		require.Equal(t, originalSize-compressedSize, gains)
	})

	t.Run("repeating group", func(t *testing.T) {
		seg := NewGroupSegment([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 0, 100, 200)
		originalSize := int(30)
		compressedSize := int(len(seg.Encode()))
		require.Equal(t, compressedSize, seg.getCompressedSize())
		gains := seg.GetCompressionGains()
		require.Equal(t, originalSize-compressedSize, gains)
	})
}
