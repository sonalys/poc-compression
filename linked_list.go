package gompressor

type ListEntry[T comparable] struct {
	Value      T
	Prev, Next *ListEntry[T]
	Ref        *LinkedList[T]
}

type LinkedList[T comparable] struct {
	Head, Tail *ListEntry[T]
	Len        int
}

func NewLinkedList[T comparable]() *LinkedList[T] {
	return &LinkedList[T]{}
}

func (l *LinkedList[T]) Find(value T) *ListEntry[T] {
	if l == nil {
		return nil
	}
	cur := l.Head
	for {
		if cur == nil {
			break
		}
		if cur.Value == value {
			return cur
		}
		cur = cur.Next
	}
	return nil
}

// Append adds a segment after the current.
func (l *LinkedList[T]) Append(next *ListEntry[T]) *LinkedList[T] {
	if next == nil {
		return l
	}
	if l.Head == nil {
		l.Head = next
	}
	if l.Tail == nil {
		l.Tail = next.Tail()
	} else {
		l.Tail.Next = next
		next.Prev = l.Tail
	}
	// Update ref inside sub-list.
	cur := next
	for {
		if cur == nil {
			break
		}
		l.Len++
		cur.Ref = l
		cur = cur.Next
	}
	return l
}

func (l *LinkedList[T]) AppendValue(value T) *LinkedList[T] {
	entry := &ListEntry[T]{
		Value: value,
		Ref:   l,
		Prev:  l.Tail,
	}
	if l.Head == nil {
		l.Head = entry
	}
	if l.Tail != nil {
		l.Tail.Next = entry
	}
	l.Tail = entry
	l.Len++
	return l
}

func (l *ListEntry[T]) Remove() {
	if l.Next != nil {
		l.Next.Prev = l.Prev
	}
	if l.Prev != nil {
		l.Prev.Next = l.Next
	}
	if l.Ref.Head == l {
		l.Ref.Head = l.Next
	}
	if l.Ref.Tail == l {
		l.Ref.Tail = l.Prev
	}
	l.Ref.Len--
}

func (l *ListEntry[T]) Find(value T) *ListEntry[T] {
	if l == nil {
		return nil
	}
	cur := l
	for {
		if cur == nil {
			break
		}
		if cur.Value == value {
			return cur
		}
	}
	return nil
}

// Append adds a segment after the current.
func (l *ListEntry[T]) Append(next *ListEntry[T]) *ListEntry[T] {
	// If next is nil do nothing, return current.
	if next == nil {
		return l
	}
	// Update ref inside sub-list.
	cur := next
	for {
		if cur == nil {
			break
		}
		l.Ref.Len++
		cur.Ref = l.Ref
		cur = cur.Next
	}
	// Update head, tail ref.
	if l.Ref.Head == nil {
		l.Ref.Head = next
	}
	if l.Ref.Tail == nil {
		l.Ref.Tail = next.Tail()
	}
	// Return last element of chain.
	if l == nil {
		return next.Tail()
	}
	// Update ref between the 2 chains.
	next.Tail().Next = l.Next
	l.Next = next
	next.Prev = l
	// Return last element of chain.
	return next.Tail()
}

func (l *ListEntry[T]) Tail() *ListEntry[T] {
	cur := l
	for {
		if cur.Next == nil {
			break
		}
		cur = cur.Next
	}
	return cur
}
