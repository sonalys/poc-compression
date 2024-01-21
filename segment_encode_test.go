package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SegmentEncoding(t *testing.T) {
	t.Run("uncompressed segment", func(t *testing.T) {
		segment := NewSegment(TypeRepeatingGroup, []byte{255, 254, 244}, 1, 2, 3)
		buffer := segment.Encode()
		got, pos := DecodeSegment(buffer)
		if pos != int(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})

	t.Run("repeat segment", func(t *testing.T) {
		segment := NewRepeatSegment(2, []byte{1, 2, 3}, 1, 2, 3)
		buffer := segment.Encode()
		got, pos := DecodeSegment(buffer)
		if pos != int(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})
}
