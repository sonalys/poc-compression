package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetCompressionGain(t *testing.T) {
	t.Run("positive gain", func(t *testing.T) {
		seg := &diskSegment{
			order: []uint8{0, 1},
			segment: &segment{
				flags:  meta(typeRepeat),
				repeat: 50,
				buffer: []byte{1, 2},
				pos:    []uint32{0, 100},
			},
		}

		compressed := seg.encodeSegment()
		originalSize := int64(200)
		compressedSize := int64(len(compressed))
		gains := seg.GetCompressionGains()

		require.Equal(t, originalSize-compressedSize, gains)
	})

	t.Run("negative gain", func(t *testing.T) {
		seg := &diskSegment{
			order: []uint8{0, 1},
			segment: &segment{
				flags:  meta(typeUncompressed),
				repeat: 1,
				buffer: []byte{1, 2},
				pos:    []uint32{0, 100},
			},
		}

		compressed := seg.encodeSegment()
		originalSize := int64(4)
		compressedSize := int64(len(compressed))
		gains := seg.GetCompressionGains()

		require.Equal(t, originalSize-compressedSize, gains)
	})
}
