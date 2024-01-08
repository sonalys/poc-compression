package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Encoding(t *testing.T) {
	t.Run("uncompressed segment", func(t *testing.T) {
		segment := orderedSegment{
			segment: &segment{
				flags:  meta(typeUncompressed),
				repeat: 1,
				buffer: []byte{1, 2, 3},
			},
			order: []uint8{0, 1, 2},
		}

		buffer := segment.encodeSegment()

		got, pos := decodeSegment(buffer)
		if pos != uint32(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})

	t.Run("repeat segment", func(t *testing.T) {
		segment := orderedSegment{
			segment: &segment{
				flags:  meta(typeRepeat),
				repeat: 2,
				buffer: []byte{1, 2, 3},
			},
			order: []uint8{0, 1, 2},
		}

		buffer := segment.encodeSegment()

		got, pos := decodeSegment(buffer)
		if pos != uint32(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})
}
