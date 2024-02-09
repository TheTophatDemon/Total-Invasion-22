package scene

type Id[T any] struct {
	Handle
}

func (id Id[T]) Get() (*T, bool) {
	return Get[*T](id.Handle)
}
