package scene

type Id[T any] struct {
	Handle
}

func (id Id[T]) Get() (T, bool) {
	return id.Handle.Get[T]()
}
