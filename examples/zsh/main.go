package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/sonalys/gompressor"
)

var paths = []string{
	"/bin/zsh",
	"/home/raicon/Downloads/snake.com",
	"/storage/DJI_0003.MP4",
	"/home/raicon/Pictures/Screenshot_20240105_145006.png",
}

func main() {
	t1 := time.Now()
	in, err := os.ReadFile(paths[0])
	if err != nil {
		panic("failed to read file")
	}
	// in = in[:math.MaxUint16]
	block := gompressor.Compress(in)
	compressedOut := gompressor.Encode(block)
	compressedSize := int(len(compressedOut))

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
	fmt.Printf(`ratio:				%.2f (%d / %d)
compressed:			%d bytes
segments:			%d
minGain:			%d bytes
maxGain:			%d bytes
took:				%s
`,
		ratio,
		compressedSize,
		int(len(in)),
		int(len(in))-compressedSize,
		segmentCount,
		minGain,
		maxGain,
		time.Since(t1),
	)
}
