package segments

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SegmentEncoding(t *testing.T) {
	t.Run("masked segment", func(t *testing.T) {
		buf := []byte{255, 254, 244}
		segment := NewMaskedSegment(buf, 1)
		buffer := segment.Encode()
		got, pos := DecodeSegment(buffer)
		if pos != int(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
		require.Equal(t, buf, got.Decompress(1))
	})

	t.Run("masked segment with zero mask and no invert", func(t *testing.T) {
		buf := []byte{0, 0, 0, 0}
		segment := NewMaskedSegment(buf, 1)
		buffer := segment.Encode()
		got, pos := DecodeSegment(buffer)
		if pos != int(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
		require.Equal(t, buf, got.Decompress(1))
	})

	t.Run("masked segment with zero mask and invert", func(t *testing.T) {
		buf := []byte{0, 0, 0, 0, 255, 255, 255}
		segment := NewMaskedSegment(buf, 1)
		buffer := segment.Encode()
		got, pos := DecodeSegment(buffer)
		if pos != int(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
		require.Equal(t, buf, got.Decompress(1))
	})

	t.Run("group segment", func(t *testing.T) {
		segment := NewGroupSegment([]byte{255, 254, 244}, 1, 2, 3)
		buffer := segment.Encode()
		got, pos := DecodeSegment(buffer)
		if pos != int(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})

	t.Run("repeat segment", func(t *testing.T) {
		segment := NewRepeatSegment(2, 1, 1, 2, 3)
		buffer := segment.Encode()
		got, pos := DecodeSegment(buffer)
		if pos != int(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})

	t.Run("repeat segment 2", func(t *testing.T) {
		posList := make([]int, 300)
		for i := range posList {
			posList[i] = i * 1000
		}
		segment := NewRepeatSegment(300, 255, posList...)
		buffer := segment.Encode()
		got, pos := DecodeSegment(buffer)
		if pos != int(len(buffer)) {
			t.Fatalf("decode returned wrong buffer position")
		}
		require.Equal(t, segment, got)
	})
}
