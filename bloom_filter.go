package gompressor

import "github.com/bits-and-blooms/bloom/v3"

type BloomFilter byte

func (b BloomFilter) Write(v byte) BloomFilter {
	return b | BloomFilter(v)
}

func (b BloomFilter) Check(v byte) bool {
	return b^BloomFilter(v) == b-BloomFilter(v)
}

func RegenerateDict(b *bloom.BloomFilter) []byte {
	out := []byte{}
	for char := 0; char < 256; char++ {
		char := byte(char)
		if !b.Test([]byte{char}) {
			continue
		}
		out = append(out, char)
	}
	return out
}
