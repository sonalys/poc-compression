package gompressor

import (
	"bytes"
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
	in = in[:400000]
	block := Compress(in)
	compressedOut := Encode(block)
	compressedSize := int64(len(compressedOut))

	t.Run("reconstruction", func(t *testing.T) {
		block2, err := Decode(compressedOut)
		require.NoError(t, err)
		out := Decompress(block2)
		require.Equal(t, len(in), len(out))
		require.True(t, bytes.Equal(in, out))
	})

	t.Run("statistics", func(t *testing.T) {
		var segmentCount int
		var minRepeat, maxRepeat int = math.MaxInt, 0
		var minGain, maxGain int64 = math.MaxInt64, 0
		var minBufferSize int64 = math.MaxInt64
		cur := block.List.Head
		for {
			if cur == nil {
				break
			}
			segmentCount++
			if bufSize := int64(len(cur.Value.Buffer)); bufSize*int64(cur.Value.Repeat) < minBufferSize {
				minBufferSize = bufSize * int64(cur.Value.Repeat)
			}
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
ratio:			%.2f (%d / %d)
compressed:	%d bytes
segments:		%d
repeat:			%d min %d max
minBuffer:	%d
minGain:		%d bytes
maxGain:		%d bytes
`,
			ratio,
			compressedSize,
			int64(len(in)),
			int64(len(in))-compressedSize,
			segmentCount,
			minRepeat,
			maxRepeat,
			minBufferSize,
			minGain,
			maxGain,
		)
	})
}
