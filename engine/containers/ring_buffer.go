package containers

type RingBuffer[T any] struct {
	buffer            []T
	head, tail, count int
}

func NewRingBuffer[T any](buffer []T) RingBuffer[T] {
	return RingBuffer[T]{
		buffer: buffer,
		head:   0,
		tail:   0,
		count:  0,
	}
}

func (rb *RingBuffer[T]) Dequeue() (out T, empty bool) {
	if rb.count == 0 {
		empty = true
		return
	}
	empty = false
	out = rb.buffer[rb.tail]
	rb.tail = (rb.tail + 1) % len(rb.buffer)
	rb.count -= 1
	return
}

func (rb *RingBuffer[T]) Enqueue(input T) bool {
	if rb.count >= len(rb.buffer)-1 {
		return false
	}
	rb.count += 1
	rb.buffer[rb.head] = input
	rb.head = (rb.head + 1) % len(rb.buffer)
	return true
}
