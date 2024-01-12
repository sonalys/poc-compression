package gompressor

import (
	"testing"

	bloom "github.com/bits-and-blooms/bloom/v3"
	"github.com/stretchr/testify/require"
)

func TestBloomFilter_Check(t *testing.T) {
	tests := []struct {
		name string
		b    BloomFilter
		v    byte
		want bool
	}{
		{
			name: "empty",
			b:    BloomFilter(0),
			v:    0,
			want: true,
		},
		{
			name: "bloom filter is set",
			b:    BloomFilter(1),
			v:    0,
			want: true,
		},
		{
			name: "value is set",
			b:    BloomFilter(0),
			v:    1,
			want: false,
		},
		{
			name: "not present",
			b:    BloomFilter(0b10),
			v:    0b11,
			want: false,
		},
		{
			name: "not present 2",
			b:    BloomFilter(0b10),
			v:    0b01,
			want: false,
		},
		{
			name: "present",
			b:    BloomFilter(0b10),
			v:    0b10,
			want: true,
		},
		{
			name: "present 2",
			b:    BloomFilter(0b101),
			v:    0b100,
			want: true,
		},
		{
			name: "present - random",
			b:    BloomFilter(0b10101101),
			v:    0b100,
			want: true,
		},
		{
			name: "not present - random",
			b:    BloomFilter(0b10101101),
			v:    0b110,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.Check(tt.v); got != tt.want {
				t.Errorf("BloomFilter.Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

func getKeySize(n int) int {
	if n >= 0b10000000 {
		return 8
	}
	if n >= 0b1000000 {
		return 7
	}
	if n >= 0b100000 {
		return 6
	}
	if n >= 0b10000 {
		return 5
	}
	if n >= 0b1000 {
		return 4
	}
	if n >= 0b100 {
		return 3
	}
	if n >= 0b10 {
		return 2
	}
	return 1
}

func Test_DictionaryRegeneration(t *testing.T) {
	t.Run("bloom test", func(t *testing.T) {
		dict := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
		filter := bloom.New(64*2, 1)
		for _, char := range dict {
			filter.Add([]byte{char})
		}
		masksize := len(filter.BitSet().Bytes()) * 8
		t.Logf("bloom filter size: %d", masksize)
		reg := RegenerateDict(filter)
		regLen := len(reg)
		t.Logf("dict[%d] = %v", regLen, reg)

		originalSizeBits := len(dict) * 8
		compressedSizeBits := len(dict)*getKeySize(regLen) + masksize
		require.Less(t, compressedSizeBits, originalSizeBits)
	})
}
