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
		require.Equal(t, compressedSize, seg.GetCompressedSize())
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
		require.Equal(t, compressedSize, seg.GetCompressedSize())
		gains := seg.GetCompressionGains()
		require.Equal(t, originalSize-compressedSize, gains)
	})

	t.Run("repeating group", func(t *testing.T) {
		seg := &Segment{
			Type:   TypeRepeatingGroup,
			Buffer: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			Pos:    []int64{0, 100, 200},
			Repeat: 1,
		}
		originalSize := int64(30)
		compressedSize := int64(len(seg.Encode()))
		require.Equal(t, compressedSize, seg.GetCompressedSize())
		gains := seg.GetCompressionGains()
		require.Equal(t, originalSize-compressedSize, gains)
	})
}
