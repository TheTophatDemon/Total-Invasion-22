package scene

type Handle struct {
	index, generation uint16
	storage           StorageOps
}

func NewHandle(index, generation uint16, storage StorageOps) Handle {
	return Handle{index, generation, storage}
}

func Get[T any](handle Handle) (T, bool) {
	var empty T
	if handle.IsNil() {
		return empty, false
	}
	data, exists := handle.storage.GetUntyped(handle)
	if !exists {
		return empty, false
	}
	typedData, isType := data.(T)
	if !isType {
		return empty, false
	}
	return typedData, true
}

func (h1 Handle) Equals(h2 Handle) bool {
	return h1.index == h2.index && h1.generation == h2.generation && h1.storage == h2.storage
}

func (h Handle) Exists() bool {
	if h.IsNil() {
		return false
	}
	return h.storage.Has(h)
}

func (h Handle) Remove() {
	if h.IsNil() {
		return
	}
	h.storage.Remove(h)
}

func (h Handle) Index() uint16 {
	return h.index
}

func (h Handle) Generation() uint16 {
	return h.generation
}

func (h Handle) IsNil() bool {
	return h.storage == nil
}
