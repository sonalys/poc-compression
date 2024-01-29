package gompressor

import (
	"bytes"

	"github.com/sonalys/gompressor/bitbuffer"
)

// var layers = []func([]byte) []byte{
// 	// segments.CreateSameCharLayer,
// 	// segments.CreateGroupSegments,
// 	segments.CreateMaskedSegments,
// }

func Compress(in []byte) *Block {
	// list := ll.NewLinkedList[segments.Segment]()
	// for _, compressionLayer := range layers {
	// 	in = compressionLayer(in)
	// }

	list := make([][]byte, 0, 256)
	indexer := make([][]int, 256)
	used := make([]bool, 0, 256)

	w := bitbuffer.NewBitBuffer(make([]byte, 0, len(in)))
	for idx := 0; idx < len(in); idx++ {
		listIdx := -1
		for _, pos := range indexer[in[idx]] {
			end := idx + len(list[pos])
			if end > len(in) {
				end = len(in)
			}
			if bytes.Equal(list[pos], in[idx:end]) {
				listIdx = pos
			}
		}
		if listIdx == -1 {
			list = append(list, []byte{in[idx]})
			indexer[in[idx]] = append(indexer[in[idx]], len(list)-1)
			used = append(used, false)
			w.Write(in[idx], 8)
			continue
		}
		if !used[listIdx] {
			used[listIdx] = true
			buf := in[idx : idx+len(list[listIdx])+1]
			list = append(list, buf)
			indexer[buf[0]] = append(indexer[buf[0]], len(list)-1)
			w.WriteBuffer(buf)
			used = append(used, false)
			idx += len(list[listIdx])
			continue
		}
		used[listIdx] = true
		w.WriteCompact(listIdx, bitbuffer.GetBitUsage(len(list)))
		idx += len(list[listIdx]) - 1
	}

	return &Block{
		Buffer: w.Buffer,
	}
}

func Decompress(b *Block) []byte {
	out := make([]byte, 0, 100000)
	r := bitbuffer.NewBitBuffer(b.Buffer)
	list := make([][]byte, 0, 256)
	indexer := make([][]int, 256)
	used := make([]bool, 0, 256)
	for idx := 0; idx < len(b.Buffer)*8; idx += 8 {
		listIdx := -1
		for _, pos := range indexer[r.Read(idx, 8)] {
			end := idx + len(list[pos])*8
			if end > len(b.Buffer)*8 {
				break
			}
			if bytes.Equal(list[pos], r.ReadBuffer(idx, len(list[pos]), 8)) {
				listIdx = pos
			}
		}
		switch {
		case listIdx == -1:
			value := r.Read(idx, 8)
			list = append(list, []byte{value})
			indexer[value] = append(indexer[value], len(list)-1)
			used = append(used, false)
			out = append(out, value)
		case !used[listIdx]:
			used[listIdx] = true
			if idx+(len(list[listIdx])+1)*8 > len(b.Buffer)*8 {
				out = append(out, list[listIdx]...)
				break
			}
			buf := r.ReadBuffer(idx, len(list[listIdx])+1, 8)
			list = append(list, buf)
			indexer[buf[0]] = append(indexer[buf[0]], len(list)-1)
			out = append(out, buf...)
			used = append(used, false)
			idx += len(list[listIdx]) * 8
		default:
			keySize := bitbuffer.GetBitUsage(len(list))
			listIdx = r.ReadCompact(idx, keySize)
			used[listIdx] = true
			out = append(out, list[listIdx]...)
			idx += keySize - 8
		}
	}
	return out
}
