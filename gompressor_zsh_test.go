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
	// in = in[:math.MaxUint16]
	block := Compress(in)
	compressedOut := Encode(block)
	compressedSize := int(len(compressedOut))

	t.Run("reconstruction", func(t *testing.T) {
		block2, err := Decode(compressedOut)
		require.NoError(t, err)

		curExp := block.List.Head
		curGot := block2.List.Head
		for {
			if curExp == nil {
				require.Nil(t, curGot)
				break
			}
			require.Equal(t, curExp.Value, curGot.Value)
			curExp = curExp.Next
			curGot = curGot.Next
		}

		out := Decompress(block2)
		require.Equal(t, len(in), len(out))
		for i := range in {
			require.Equal(t, in[i], out[i], i)
		}
		require.True(t, bytes.Equal(in, out))
	})

	t.Run("statistics", func(t *testing.T) {
		var segmentCount int
		var minGain, maxGain int = math.MaxInt, 0
		cur := block.List.Head
		for {
			if cur == nil {
				break
			}
			segmentCount++
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
minGain:		%d bytes
maxGain:		%d bytes
`,
			ratio,
			compressedSize,
			int(len(in)),
			int(len(in))-compressedSize,
			segmentCount,
			minGain,
			maxGain,
		)
	})
}

func Test_chunksZSH(t *testing.T) {
	in, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	var compressedSize int
	chunkSize := math.MaxUint16
	for i := 0; i < len(in); i += chunkSize {
		end := i + chunkSize
		if end > len(in) {
			end = len(in)
		}
		buf := in[i:end]
		block := Compress(buf)
		compressedOut := Encode(block)
		compressedSize += int(len(compressedOut))
	}
	t.Logf("ratio: %.2f (%d / %d)", float64(compressedSize)/float64(len(in)), compressedSize, len(in))
}
