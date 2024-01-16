package gompressor

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var paths = []string{
	"/bin/zsh",
	"/home/raicon/Downloads/snake.com",
	"/storage/DJI_0003.MP4",
	"/home/raicon/Pictures/Screenshot_20240105_145006.png",
}

func Test_compressZSH(t *testing.T) {
	in, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}

	block := NewBlock(in)
	compressedOut := Compress(block)
	compressedSize := int64(len(compressedOut))

	t.Run("test decompression", func(t *testing.T) {
		out := Decompress(block)
		require.Equal(t, len(in), len(out), "input and output are different")
		for i := range in {
			if in[i] != out[i] {
				window := 20
				require.Equal(t, in[i], out[i], "in[%d] != out[%d]\n%v\n%v", i, i, in[i:i+window], out[i:i+window])
			}
		}
	})

	var segmentCount int
	var minRepeat, maxRepeat int = math.MaxInt, 0
	var minGain, maxGain int64 = math.MaxInt64, 0

	cur := block.List.Head
	for {
		if cur == nil {
			break
		}
		segmentCount++
		if repeat := int(cur.Value.Repeat); repeat > maxRepeat {
			maxRepeat = repeat
		} else if repeat < minRepeat {
			minRepeat = repeat
		}
		if gain := cur.Value.GetCompressionGains(); gain > maxGain {
			maxGain = gain
		} else if gain < minGain {
			minGain = gain
		}
		cur = cur.Next
	}

	ratio := float64(compressedSize) / float64(len(in))
	t.Logf(`
byte ratio				%.2f (%d / %d)
compressed: 			%d bytes
segments count: 	%d
min repeat: 			%d
max repeat:			 	%d
min gain:				 	%d bytes
max gain: 				%d bytes
`,
		ratio,
		compressedSize,
		int64(len(in)),
		int64(len(in))-compressedSize,
		segmentCount,
		minRepeat,
		maxRepeat,
		minGain,
		maxGain,
	)
}
