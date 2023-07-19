package containers

// A linked list that uses generics.
type List[T comparable] struct {
	front *Element[T]
	back  *Element[T]
	size  int
}

func NewList[T comparable]() *List[T] {
	return &List[T]{
		front: nil,
		back:  nil,
		size:  0,
	}
}

func (l *List[T]) IsEmpty() bool {
	return l.size == 0
}

func (l *List[T]) Size() int {
	return l.size
}

func (l *List[T]) Front() *Element[T] {
	return l.front
}

func (l *List[T]) Back() *Element[T] {
	return l.back
}

func (l *List[T]) PushBack(value T) *Element[T] {
	elem := &Element[T]{
		Value: value,
		next:  nil,
		prev:  l.back,
	}
	if l.front == nil {
		l.front = elem
		l.back = elem
	} else if l.back != nil {
		l.back.next = elem
		l.back = elem
	}
	l.size += 1
	return elem
}

func (l *List[T]) RemoveElem(e *Element[T]) {
	before := e.prev
	after := e.next
	if before != nil {
		before.next = after
	} else {
		l.front = after
	}
	if after != nil {
		after.prev = before
	} else {
		l.back = before
	}
	l.size -= 1
}

func (l *List[T]) FindElem(value T) *Element[T] {
	for it := l.front; it != nil; it = it.next {
		if it.Value == value {
			return it
		}
	}
	return nil
}

func (l *List[T]) Remove(value T) bool {
	if it := l.FindElem(value); it != nil {
		l.RemoveElem(it)
		return true
	}
	return false
}

type Element[T any] struct {
	Value T
	next  *Element[T]
	prev  *Element[T]
}

func (e *Element[T]) Next() *Element[T] {
	return e.next
}

func (e *Element[T]) Prev() *Element[T] {
	return e.prev
}
