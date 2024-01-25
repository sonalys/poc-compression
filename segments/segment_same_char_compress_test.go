package segments

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_bitWriter(t *testing.T) {
	buf := make([]byte, 0, 10)
	w := newBitWriter(buf)
	var value byte = 7
	w.WriteByte(value, 7)
	require.EqualValues(t, value<<1, w.buffer[0])
	value = 127
	w.WriteByte(value, 7)
	require.EqualValues(t, 7<<1+1, w.buffer[0])
	require.EqualValues(t, value<<2, w.buffer[1])
	value = 85
	prev := w.buffer[1]
	w.WriteByte(value, 7)
	require.EqualValues(t, prev+value>>5, w.buffer[1])
	require.EqualValues(t, value<<3, w.buffer[2])
}
