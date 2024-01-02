package main

import (
	"bytes"
	"unsafe"
)

type repetitionEntry struct {
	Position uint32
	Size     uint32
	Byte     byte
}

type repetitionDictionary struct {
	Size    uint32
	Entries []repetitionEntry
}

func countBytes(input []byte) (repetition []uint) {
	repetition = make([]uint, 256)
	for _, char := range input {
		repetition[char] = repetition[char] + 1
	}
	return
}

var entrySize = uint32(unsafe.Sizeof(repetitionEntry{}))

func compressRepetition(in []byte, minSize uint32) (entries []repetitionEntry, out []byte) {
	out = make([]byte, 0, len(in))
	for i := uint32(0); i < uint32(len(in)); i++ {
		size := uint32(1)
		for j := i + 1; j < uint32(len(in)) && in[i] == in[j]; j++ {
			size++
		}
		// ensure compression is net positive.
		if size > minSize {
			entries = append(entries, repetitionEntry{
				Position: uint32(len(out)),
				Size:     size,
				Byte:     in[i],
			})
			i += size - 1
		} else {
			out = append(out, in[i])
		}
	}
	return
}

func reconstruct(in []byte, entries []repetitionEntry) (out []byte) {
	out = make([]byte, 0, len(in))
	prevPos := 0
	for _, entry := range entries {
		curPos := int(entry.Position)
		out = append(out, in[prevPos:curPos]...)
		out = append(out, bytes.Repeat([]byte{entry.Byte}, int(entry.Size))...)
		prevPos = curPos
	}
	out = append(out, in[prevPos:]...)
	return
}
