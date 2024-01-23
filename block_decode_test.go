package gompressor

import (
	"testing"

	ll "github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
	"github.com/stretchr/testify/require"
)

func Test_BlockEncoding(t *testing.T) {
	b := &Block{
		OriginalSize: 100,
		Segments:     &ll.LinkedList[segments.Segment]{},
		Buffer:       []byte{1, 2, 3},
	}
	encoded := Encode(b)
	got, err := Decode(encoded)
	require.NoError(t, err)
	require.Equal(t, b, got)
}
