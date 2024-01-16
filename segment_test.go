package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetCompressionGain(t *testing.T) {
	t.Run("positive gain", func(t *testing.T) {
		seg := &Segment{
			Type:   TypeRepeatSameChar,
			Buffer: []byte{1, 2},
			Pos:    []int64{0, 100},
			Repeat: 50,
		}
		originalSize := int64(200)
		compressedSize := int64(len(seg.Encode()))
		gains := seg.GetCompressionGains()
		require.Equal(t, originalSize-compressedSize, gains)
	})

	t.Run("negative gain", func(t *testing.T) {
		seg := &Segment{
			Type:   TypeUncompressed,
			Buffer: []byte{1, 2},
			Pos:    []int64{0, 100},
		}
		originalSize := int64(4)
		compressedSize := int64(len(seg.Encode()))
		gains := seg.GetCompressionGains()
		require.Equal(t, originalSize-compressedSize, gains)
	})
}
