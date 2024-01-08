package gompressor

func getOrderedDecompressionList(list []DiskSegment) (out []DiskSegment) {
	out = make([]DiskSegment, 0, len(list))
	var order uint16 = 0
	for {
		found := false
		for _, entry := range list {
			for _, curOrder := range entry.Order {
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
	out := make([]byte, 0, in.Size)
	for _, entry := range getOrderedDecompressionList(in.Segments) {
		out = append(out, entry.Decompress()...)
	}
	return out
}
