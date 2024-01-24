package segments

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateRepeatingSegments(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		buf := []byte{}
		list, uncompressed := CreateGroupSegments(buf)
		require.Zero(t, list.Len)
		require.Empty(t, uncompressed)
	})
	t.Run("no repeating", func(t *testing.T) {
		buf := []byte{1, 2, 3, 4, 5, 6}
		list, uncompressed := CreateGroupSegments(buf)
		require.Zero(t, list.Len)
		require.Equal(t, buf, uncompressed)
	})
	t.Run("small repeating, startOffset = 0, endOffset = 2", func(t *testing.T) {
		buf := []byte{1, 2, 3, 4, 1, 2, 3, 4, 1, 2, 3, 4}
		list, uncompressed := CreateGroupSegments(buf)
		require.Empty(t, uncompressed)
		require.Equal(t, 1, list.Len)
		seg := list.Head.Value
		require.Equal(t, []byte{1, 2, 3, 4}, seg.Decompress(1))
		require.Equal(t, []int{0, 4, 8}, seg.GetPos())
	})
	t.Run("single repeating pattern, no overlap", func(t *testing.T) {
		buf := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		list, uncompressed := CreateGroupSegments(buf)
		require.Empty(t, uncompressed)
		require.Equal(t, 1, list.Len)
		seg := list.Head.Value
		require.Equal(t, []byte{0, 0, 0, 0}, seg.Decompress(1))
		require.Equal(t, []int{0, 4, 8}, seg.GetPos())
	})
	t.Run("overlap should not be an issue", func(t *testing.T) {
		buf := []byte{1, 2, 3, 2, 1, 2, 3, 2, 1, 2, 3, 2, 1}
		list, uncompressed := CreateGroupSegments(buf)
		require.Equal(t, []byte{1}, uncompressed)
		require.Equal(t, 1, list.Len)
		seg := list.Head.Value
		require.Equal(t, []byte{1, 2, 3, 2}, seg.Decompress(1))
		require.Equal(t, []int{0, 4, 8}, seg.GetPos())
	})
	t.Run("overlap should not be an issue 2", func(t *testing.T) {
		buf := []byte{0, 0, 1, 0, 0, 1, 0, 0, 1, 0, 0, 1}
		list, uncompressed := CreateGroupSegments(buf)
		require.Equal(t, 1, list.Len)
		seg := list.Head.Value
		require.Equal(t, []byte{1, 0, 0, 1}, seg.Decompress(1))
		require.Equal(t, []int{2, 8}, seg.GetPos())
		require.Equal(t, []byte{0, 0, 0, 0}, uncompressed)
	})
}
