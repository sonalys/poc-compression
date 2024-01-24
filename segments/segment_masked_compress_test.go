package segments

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_CreateMaskedSegments(t *testing.T) {
	in := []byte{255, 255, 0, 0, 0, 0, 0, 0}
	list, out := CreateMaskedSegments(in)
	require.Empty(t, out)
	require.Equal(t, list.Len, 1)
}
