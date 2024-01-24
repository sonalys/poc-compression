package linkedlist

type ListEntry[T any] struct {
	Value      T
	Prev, Next *ListEntry[T]
	Ref        *LinkedList[T]
}

type LinkedList[T any] struct {
	Head, Tail *ListEntry[T]
	Len        int
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

func (l *LinkedList[T]) AppendValue(value T) *LinkedList[T] {
	entry := &ListEntry[T]{
		Value: value,
		Ref:   l,
		Prev:  l.Tail,
	}
	return l.Append(entry)
}

// Append adds a segment after the current.
func (l *LinkedList[T]) Append(entry *ListEntry[T]) *LinkedList[T] {
	if entry == nil {
		return l
	}
	if l.Head == nil {
		l.Head = entry
	}
	if l.Tail != nil {
		l.Tail.Next = entry
		entry.Prev = l.Tail
	}
	// Update ref inside sub-list.
	cur := entry
	for {
		l.Len++
		cur.Ref = l
		next := cur.Next
		if next == nil {
			l.Tail = cur
			break
		}
		cur = next
	}
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
