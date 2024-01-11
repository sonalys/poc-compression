package gompressor

type Block struct {
	Size   uint32
	Head   *Segment
	Buffer []byte
}

// Remove dereferences this segment from the linked list.
func (b *Block) Remove(s *Segment) {
	if s.Previous == nil {
		b.Head = s.Next
	} else {
		s.Previous.Next = s.Next
	}
	if s.Next != nil {
		s.Next.Previous = s.Previous
	}
}

// Remove deletes element.
func (s *Segment) Remove() *Segment {
	if s.Next != nil {
		s.Next.Previous = s.Previous
	}
	if s.Previous != nil {
		s.Previous.Next = s.Next
	}
	return s.Next
}

func (s *Segment) ForEach(f func(*Segment)) {
	cur := s
	for {
		if cur == nil {
			break
		}
		f(cur)
		cur = cur.Next
	}
}
