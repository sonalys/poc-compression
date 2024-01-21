package gompressor

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateRepeatingSegments(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		buf := []byte{}
		list, uncompressed := CreateRepeatingSegments(buf)
		require.Zero(t, list.Len)
		require.Empty(t, uncompressed)
	})
	t.Run("no repeating", func(t *testing.T) {
		buf := []byte{1, 2, 3, 4, 5, 6}
		list, uncompressed := CreateRepeatingSegments(buf)
		require.Zero(t, list.Len)
		require.Equal(t, buf, uncompressed)
	})
	t.Run("small repeating, startOffset = 0, endOffset = 2", func(t *testing.T) {
		buf := []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
		list, uncompressed := CreateRepeatingSegments(buf)
		require.Empty(t, uncompressed)
		require.Equal(t, 1, list.Len)
		seg := list.Head.Value
		require.Equal(t, []byte{1, 2, 3, 4}, seg.Buffer)
		require.Equal(t, []int{0, 4, 8}, seg.Pos)
	})
	t.Run("single repeating pattern, no overlap", func(t *testing.T) {
		buf := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		list, uncompressed := CreateRepeatingSegments(buf)
		require.Empty(t, uncompressed)
		require.Equal(t, 1, list.Len)
		seg := list.Head.Value
		require.Equal(t, []byte{0, 0}, seg.Buffer)
		require.Equal(t, []int{0, 2, 4, 6, 8, 10}, seg.Pos)
	})
	t.Run("overlap should not be an issue", func(t *testing.T) {
		buf := []byte{1, 2, 3, 2, 1, 2, 3, 2, 1, 2, 3, 2, 1}
		list, uncompressed := CreateRepeatingSegments(buf)
		require.Equal(t, []byte{1}, uncompressed)
		require.Equal(t, 1, list.Len)
		seg := list.Head.Value
		require.Equal(t, []byte{2, 3, 2, 1}, seg.Buffer)
		require.Equal(t, []int{1, 5, 9}, seg.Pos)
	})

	t.Run("overlap should not be an issue 2", func(t *testing.T) {
		buf := []byte{0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1}
		list, uncompressed := CreateRepeatingSegments(buf)
		require.Empty(t, uncompressed)
		require.Equal(t, 1, list.Len)
		seg := list.Head.Value
		require.Equal(t, []byte{0, 0, 1}, seg.Buffer)
		require.Equal(t, []int{0, 3, 6, 9}, seg.Pos)
	})
}
