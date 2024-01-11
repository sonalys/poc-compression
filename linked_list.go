package gompressor

type ListEntry[T any] struct {
	Value      *T
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

func (l *LinkedList[T]) AppendValue(value *T) *LinkedList[T] {
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
}

// Append adds a segment after the current.
func (l *ListEntry[T]) Append(next *ListEntry[T]) *ListEntry[T] {
	next.Tail().Next = l.Next
	l.Next = next
	next.Prev = l
	return next
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
