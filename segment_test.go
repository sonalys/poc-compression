package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_GetCompressionGain(t *testing.T) {
	t.Run("positive gain", func(t *testing.T) {
		seg := &Segment{
			Metadata: meta(typeRepeat),
			Repeat:   50,
			Buffer:   []byte{1, 2},
			Pos:      []uint32{0, 100},
		}

		compressed := seg.Encode()
		originalSize := int64(200)
		compressedSize := int64(len(compressed))
		gains := seg.GetCompressionGains()

		require.Equal(t, originalSize-compressedSize, gains)
	})

	t.Run("negative gain", func(t *testing.T) {
		seg := &Segment{
			Metadata: meta(typeUncompressed),
			Repeat:   1,
			Buffer:   []byte{1, 2},
			Pos:      []uint32{0, 100},
		}

		compressed := seg.Encode()
		originalSize := int64(4)
		compressedSize := int64(len(compressed))
		gains := seg.GetCompressionGains()

		require.Equal(t, originalSize-compressedSize, gains)
	})
}
