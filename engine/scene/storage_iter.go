package scene

// Contains state for iterating through a storage object.
// Go's standard iterators are not used because they are
// currently too slow and require excessive heap allocations.
type StorageIter[T any] struct {
	storage *Storage[T]
	index   int
}

// Returns the next active entity and its handle from storage.
// When the end of the storage is reached, zero values are returned.
func (iter *StorageIter[T]) Next() (*T, Handle) {
	if iter == nil || iter.storage == nil || len(iter.storage.data) == 0 {
		return nil, Handle{}
	}
	iter.index++
	for ; iter.index <= iter.storage.lastActive; iter.index++ {
		if iter.storage.active[iter.index] {
			return &iter.storage.data[iter.index], iter.storage.owners[iter.index]
		}
	}
	return nil, Handle{}
}

// Returns true if calling Next() on this iterator will give another item.
func (iter *StorageIter[T]) HasNext() bool {
	if iter == nil || iter.storage == nil || len(iter.storage.data) == 0 {
		return false
	}
	return iter.index < iter.storage.lastActive
}

// Returns the iterator's state to the beginning of its storage so it can be iterated again.
func (iter *StorageIter[T]) Reset() {
	if iter == nil {
		return
	}
	iter.index = -1
}
