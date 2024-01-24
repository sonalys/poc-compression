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
		list.AppendValue(segments.NewMaskedSegment(segments.WithBuffer([]byte{4, 5, 6}), 2))
		list.AppendValue(segments.NewGroupSegment([]byte{1, 2, 3}, 1))
		list.AppendValue(segments.NewRepeatSegment(3, 1, 5))
		block := &Block{
			OriginalSize: 13,
			Segments:     list,
			Buffer:       []byte{255, 255, 255, 255},
		}
		got := Decompress(block)
		require.Equal(t, []byte{255, 1, 2, 3, 255, 1, 1, 1, 4, 5, 6, 255, 255}, got)
	})
}
