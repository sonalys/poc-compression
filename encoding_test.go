package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Encoding(t *testing.T) {
	t.Run("uncompressed segment", func(t *testing.T) {
		segment := DiskSegment{
			Segment: &Segment{
				Metadata: meta(typeUncompressed),
				Repeat:   1,
				Buffer:   []byte{1, 2, 3},
			},
			Order: []uint16{0, 1, 2},
		}

		buffer := segment.Encode()

		got, pos := DecodeSegment(buffer)
		if pos != uint32(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})

	t.Run("repeat segment", func(t *testing.T) {
		segment := DiskSegment{
			Segment: &Segment{
				Metadata: meta(typeRepeat),
				Repeat:   2,
				Buffer:   []byte{1, 2, 3},
			},
			Order: []uint16{0, 1, 2},
		}

		buffer := segment.Encode()

		got, pos := DecodeSegment(buffer)
		if pos != uint32(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})
}
