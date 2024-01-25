package bitbuffer

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzBitBuffer(f *testing.F) {
	for i := 1; i <= 8; i++ {
		for j := 0; j < 256; j++ {
			f.Add(i, byte(j))
		}
	}
	f.Fuzz(func(t *testing.T, size int, value byte) {
		value = value << (8 - size) >> (8 - size)
		b := NewBitBuffer(make([]byte, 0, 10))
		value = b.Write(value, size)
		require.EqualValues(t, value, b.Read(0*size, size))
		value = b.Write(value, size)
		require.EqualValues(t, value, b.Read(1*size, size))
		value = b.Write(value, size)
		require.EqualValues(t, value, b.Read(2*size, size))
	})
}
