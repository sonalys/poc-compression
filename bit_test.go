package gompressor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecompressByte(t *testing.T) {
	var mask byte = 0b00101100
	compressedBits := GetMaskBits(mask)
	var exp byte = 0b1100
	compressed := CompressByte(compressedBits, exp)
	t.Logf("compressed %#b to %#b", exp, compressed)
	got := DecompressByte(compressedBits, compressed)
	t.Logf("decompressed %#b to %#b", compressed, got)

	if exp != got {
		require.Failf(t, "values are not equal", "exp: %#b got: %#b", exp, got)
	}
}

func Test_DecompressBuffer(t *testing.T) {
	t.Run("small values", func(t *testing.T) {
		exp := []byte{1, 2, 3, 4, 5, 6, 7}
		mask, invert, compressed := CompressBuffer(exp)
		got := DecompressBuffer(mask, invert, compressed, len(exp))
		require.Equal(t, exp, got)
	})
	t.Run("bigger values", func(t *testing.T) {
		exp := []byte{255, 254, 253, 252, 251, 250, 249}
		mask, invert, compressed := CompressBuffer(exp)
		got := DecompressBuffer(mask, invert, compressed, len(exp))
		require.Equal(t, exp, got)
	})
}

func Test_CompressDecompressByte(t *testing.T) {
	t.Run("single byte", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			exp := []byte{byte(i)}
			mask, invert, compressed := CompressBuffer(exp)
			got := DecompressBuffer(mask, invert, compressed, 1)
			require.Equal(t, exp, got, i)
		}
	})
	t.Run("2 bytes", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			exp := bytes.Repeat([]byte{byte(i)}, 2)
			mask, invert, compressed := CompressBuffer(exp)
			got := DecompressBuffer(mask, invert, compressed, 2)
			require.Equal(t, exp, got, i)
		}
	})
	t.Run("3 bytes", func(t *testing.T) {
		for i := 0; i < 256; i++ {
			exp := []byte{byte(i), byte(i + 1), byte(i + 2)}
			mask, invert, compressed := CompressBuffer(exp)
			got := DecompressBuffer(mask, invert, compressed, 3)
			require.Equal(t, exp, got, i)
		}
	})
}
