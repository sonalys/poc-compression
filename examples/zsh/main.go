package main

import (
	"math"
	"os"
	"time"

	"github.com/sonalys/gompressor"
	ll "github.com/sonalys/gompressor/linkedlist"
	"github.com/sonalys/gompressor/segments"
	"github.com/sonalys/gompressor/utils"
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
	// in = in[:10000]
	var compressedSize int
	var uncompressedSize int
	allChunksList := ll.NewLinkedList[segments.Segment]()
	const chunkSize = math.MaxUint16
	for i := 0; i < len(in); i += chunkSize {
		end := i + chunkSize
		if end > len(in) {
			end = len(in)
		}
		chunk := in[i:end]
		block := gompressor.Compress(chunk)
		compressedOut := gompressor.Encode(block)
		compressedSize += len(compressedOut)
		uncompressedSize += len(block.Buffer)
		out := gompressor.Decompress(block)
		if err := utils.IntegrityCheck(chunk, out); err != nil {
			panic(err.Error())
		}
		allChunksList.Append(block.Segments.Head)
	}
	utils.PrintStatistics(in, compressedSize, uncompressedSize, allChunksList, t1)
}
