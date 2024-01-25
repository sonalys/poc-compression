package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/sonalys/gompressor"
	ll "github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
)

var paths = []string{
	"/bin/zsh",
	"/home/raicon/Downloads/snake.com",
	"/storage/DJI_0003.MP4",
	"/home/raicon/Pictures/Screenshot_20240105_145006.png",
}

func printStatistics(in []byte, compressedSize, uncompressedSize int, list *ll.LinkedList[segments.Segment], t1 time.Time) {
	var segmentCount int
	var minGain, maxGain int = math.MaxInt, 0
	cur := list.Head

	typeCount := map[segments.SegmentType]int{}
	typeGain := map[segments.SegmentType]int{}

	for {
		if cur == nil {
			break
		}
		segmentCount++

		gain := cur.Value.GetCompressionGains()
		t := cur.Value.GetType()

		typeCount[t] += len(cur.Value.GetPos())
		typeGain[t] += gain

		if gain > maxGain {
			maxGain = gain
		} else if gain < minGain {
			minGain = gain
		}
		cur = cur.Next
	}

	ratio := float64(compressedSize) / float64(len(in))
	fmt.Printf(`ratio:				%.2f (%d / %d)
compressed:			%d bytes
uncompressed:			%d bytes
segments storage:		%d bytes
segments:			%d
minGain:			%d bytes
maxGain:			%d bytes
took:				%s
`,
		ratio,
		compressedSize,
		int(len(in)),
		int(len(in))-compressedSize,
		uncompressedSize,
		len(in)-uncompressedSize,
		segmentCount,
		minGain,
		maxGain,
		time.Since(t1),
	)

	fmt.Println("\nType Pos count:")
	for t, count := range typeCount {
		fmt.Printf("%s:\t%d\n", segments.TypeName[t], count)
	}

	fmt.Println("\nType Gain:")
	for t, gain := range typeGain {
		fmt.Printf("%s:\t%d\tbytes\n", segments.TypeName[t], gain)
	}
}

func verifyIntegrity(in, out []byte) {
	if len(in) != len(out) {
		// TODO: figure out why it doesn't work for 3 layers, but work for any 2 layers.
		msg := fmt.Sprintf("output size is different. exp %d != got %d", len(in), len(out))
		panic(msg)
	}
	for i := range in {
		if in[i] != out[i] {
			msg := fmt.Sprintf("output is different at pos %d exp %d != got %d", i, in[i], out[i])
			panic(msg)
		}
	}
}

func main() {
	t1 := time.Now()
	in, err := os.ReadFile(paths[0])
	if err != nil {
		panic("failed to read file")
	}
	// in = in[:10000]
	var compressedSize int
	var uncompressedSize int
	allChunksList := ll.NewLinkedList[segments.Segment]()
	// const chunkSize = math.MaxUint16
	// for i := 0; i < len(in); i += chunkSize {
	// 	end := i + chunkSize
	// 	if end > len(in) {
	// 		end = len(in)
	// 	}
	chunk := in
	block := gompressor.Compress(chunk)
	compressedOut := gompressor.Encode(block)
	compressedSize += len(compressedOut)
	uncompressedSize += len(block.Buffer)
	// out := gompressor.Decompress(block)
	// verifyIntegrity(chunk, out)
	// allChunksList.Append(block.Segments.Head)
	// }
	printStatistics(in, compressedSize, uncompressedSize, allChunksList, t1)
}
