package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SegmentEncoding(t *testing.T) {
	t.Run("uncompressed segment", func(t *testing.T) {
		segment := &Segment{
			Type:   TypeUncompressed,
			Buffer: []byte{1, 2, 3},
			Pos:    []int64{1, 2, 3},
		}
		buffer := segment.Encode()

		got, pos := DecodeSegment(buffer)
		if pos != int64(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})

	t.Run("repeat segment", func(t *testing.T) {
		segment := &Segment{
			Type:   TypeRepeatSameChar,
			Buffer: []byte{1, 2, 3},
			Pos:    []int64{1, 2, 3},
			Repeat: 2,
		}

		buffer := segment.Encode()

		got, pos := DecodeSegment(buffer)
		if pos != int64(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})
}