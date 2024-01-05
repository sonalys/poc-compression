package gompressor

func compress(in []byte, minSize uint16) *block {
	if minSize == 1 {
		panic("don't be retarded")
	}
	b := block{
		size: uint32(len(in)),
	}
	// prev is a cursor to the last position before a repeating group
	var prev uint32
	head := &segment{}
	// cur is a cursor to the head of the block's segments.
	cur := head
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
			compressed, gain := newSegment(typeRepeat, index, repeatCount, []byte{in[index]})
			if gain < 0 {
				continue
			}
			// avoid creating segments with nil buffer.
			if index-prev > 0 {
				raw, gain2 := newSegment(typeUncompressed, prev, 1, in[prev:index])
				if gain+gain2 < 0 {
					continue
				}
				cur = cur.add(raw)
			}
			cur = cur.add(compressed)
			// mark the next byte as the begin of the next unsegmented section.
			prev = index + uint32(repeatCount)
			// prev -1 because the for iterator will add +1 again.
			index += prev - 1
		}
	}
	head = head.next
	if head == nil {
		head, _ = newSegment(typeUncompressed, 0, 1, in)
	} else if uint32(len(in))-prev > 0 {
		raw, _ := newSegment(typeUncompressed, prev, 1, in[prev:])
		cur.next = raw
		raw.previous = cur
	}

	head.deduplicate()

	b.head = getOrderedSegments(head)

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
	return &b
}
