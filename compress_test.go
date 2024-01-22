package gompressor

import (
	"testing"

	ll "github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
	"github.com/stretchr/testify/require"
)

func Test_Compression(t *testing.T) {
	t.Run("test 3 segments", func(t *testing.T) {
		list := ll.NewLinkedList[segments.Segment]()
		list.AppendValue(segments.NewMaskedSegment(segments.WithBuffer([]byte{4, 5, 6}), 0))
		list.AppendValue(segments.NewGroupSegment([]byte{1, 2, 3}, 2))
		list.AppendValue(segments.NewRepeatSegment(3, 1, 5))
		block := &Block{
			OriginalSize: 10,
			List:         list,
			Buffer:       []byte{255},
		}
		got := Decompress(block)
		require.Equal(t, []byte{4, 5, 1, 2, 3, 1, 1, 1, 6, 255}, got)
	})
}
