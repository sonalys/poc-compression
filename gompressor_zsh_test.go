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
	// zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// log.Logger = log.Output(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
	// 	w.NoColor = true
	// 	w.FieldsExclude = append(w.FieldsExclude, "time")
	// }))
	in, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatalf("failed to read file: %s", err)
	}
	// in = in[:100000]
	block := Compress(in)
	compressedOut := Encode(block)
	compressedSize := int(len(compressedOut))

	t.Run("reconstruction", func(t *testing.T) {
		block2, err := Decode(compressedOut)
		require.NoError(t, err)
		out := Decompress(block2)
		require.Equal(t, len(in), len(out))
		for i := range in {
			require.Equal(t, in[i], out[i], i)
		}
		require.True(t, bytes.Equal(in, out))
	})

	t.Run("statistics", func(t *testing.T) {
		var segmentCount int
		var minRepeat, maxRepeat int = math.MaxInt, 0
		var minGain, maxGain int = math.MaxInt, 0
		var minBufferSize int = math.MaxInt
		cur := block.List.Head
		for {
			if cur == nil {
				break
			}
			segmentCount++
			if bufSize := int(len(cur.Value.Buffer)); bufSize*int(cur.Value.Repeat) < minBufferSize {
				minBufferSize = bufSize * int(cur.Value.Repeat)
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
			int(len(in)),
			int(len(in))-compressedSize,
			segmentCount,
			minRepeat,
			maxRepeat,
			minBufferSize,
			minGain,
			maxGain,
		)
	})
}
