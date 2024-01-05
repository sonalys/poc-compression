package gompressor

func compress(in []byte, minSize uint16) block {
	b := block{
		size: uint32(len(in)),
		head: &segment{},
	}
	// prev is a cursor to the last position before a repeating group
	var prev uint32
	// cur is a cursor to the head of the block's segments.
	cur := b.head
	// finds repetition groups and store them.
	for index := uint32(0); index < uint32(len(in)); index++ {
		repeatCount := uint16(1)
		for j := index + 1; j < uint32(len(in)) && in[index] == in[j]; j++ {
			repeatCount += 1
			if repeatCount == 0 {
				panic("repeat overflow")
			}
		}
		if repeatCount >= minSize {
			// avoid creating segments with nil buffer.
			if index-prev > 0 {
				cur = cur.addNext(typeUncompressed, prev, 1, in[prev:index])
			}
			cur = cur.addNext(typeRepeat, index, repeatCount, []byte{in[index]})
			index += uint32(repeatCount) - 1
			prev = index
		}
	}
	b.head = b.head.next
	if b.head == nil {
		b.head = newSegment(typeUncompressed, 0, 1, in)
	} else {
		cur.addNext(typeUncompressed, prev, 1, in[prev:])
	}

	b.head.deduplicate()

	// for i := uint32(0); i < uint32(len(out)); i++ {
	// 	localGroups := []repetitionGroup{}
	// 	maxSize := uint32(0)
	// outer:
	// 	for j := i + minSize; j < uint32(len(in)); j++ {
	// 		size := uint32(minSize)
	// 		// finds the minimum size that matches i and j.
	// 		for ; size < j-i; size++ {
	// 			if out[i+size] != out[j+size] {
	// 				break
	// 			}
	// 		}
	// 		if size < minSize {
	// 			continue
	// 		}
	// 		if size > maxSize {
	// 			maxSize = size
	// 		}
	// 		for k := range groups {
	// 			if bytes.Equal(groups[k].Bytes, out[i:i+size]) {
	// 				// already registered previously.
	// 				continue outer
	// 			}
	// 		}
	// 		// if there is already a group for i with same size, then bytes is equal as well.
	// 		for k := range localGroups {
	// 			if uint32(len(localGroups[k].Bytes)) == size {
	// 				localGroups[k].Positions = append(localGroups[k].Positions, j)
	// 			}
	// 			continue outer
	// 		}
	// 		// if there is no local group, then create one for i, j.
	// 		localGroups = append(localGroups, repetitionGroup{
	// 			Positions: []uint32{i, j},
	// 			Bytes:     out[i : i+size],
	// 		})
	// 	}
	// 	if len(localGroups) == 0 {
	// 		out2 = append(out2, out[i])
	// 		continue
	// 	}
	// 	i += maxSize
	// 	groups = append(groups, localGroups...)
	// }
	// out = out2
	return b
}
