package gompressor

func Decode(in []byte) (out *Block, err error) {
	lenIn := len(in)
	var pos int
	out = &Block{
		OriginalSize: int(decoder.Uint64(in)),
		List:         &LinkedList[*Segment]{},
	}
	bufLen := int(decoder.Uint64(in[8:]))
	pos += 16
	out.Buffer, pos = in[pos:pos+bufLen], pos+bufLen
	for {
		if pos == lenIn {
			break
		}
		if pos > lenIn {
			panic("you messed up pos")
		}
		decoded, offset := DecodeSegment(in[pos:])
		pos += offset
		out.List.AppendValue(decoded)
	}
	return
}
