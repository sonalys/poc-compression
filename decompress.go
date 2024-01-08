package gompressor

func getOrderedDecompressionList(list []diskSegment) (out []diskSegment) {
	out = make([]diskSegment, 0, len(list))
	var order uint8 = 0
	for {
		found := false
		for _, entry := range list {
			for _, curOrder := range entry.order {
				if curOrder == order {
					out = append(out, entry)
					found = true
					order++
				}
			}
		}
		if !found {
			return
		}
	}
}

func Decompress(in *block) []byte {
	out := make([]byte, 0, in.size)
	for _, entry := range getOrderedDecompressionList(in.segments) {
		out = append(out, entry.Decompress()...)
	}
	return out
}
