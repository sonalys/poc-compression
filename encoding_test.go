package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SegmentEncoding(t *testing.T) {
	t.Run("uncompressed segment", func(t *testing.T) {
		segment := Segment{
			Type:   TypeUncompressed,
			Repeat: 1,
			Buffer: []byte{1, 2, 3},
			Pos:    []uint32{1, 2, 3},
		}
		buffer := segment.Encode()

		got, pos := DecodeSegment(buffer)
		if pos != uint32(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})

	t.Run("repeat segment", func(t *testing.T) {
		segment := Segment{
			Type:   TypeRepeatSameChar,
			Repeat: 2,
			Buffer: []byte{1, 2, 3},
			Pos:    []uint32{1, 2, 3},
		}

		buffer := segment.Encode()

		got, pos := DecodeSegment(buffer)
		if pos != uint32(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})
}

func Test_BlockEncoding(t *testing.T) {
	b := &Block{
		Size:   100,
		Head:   nil,
		Buffer: []byte{1, 2, 3},
	}
	encoded := Encode(b)
	got, err := Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, b, got)
}
