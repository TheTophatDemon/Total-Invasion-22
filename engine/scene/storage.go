package scene

import (
	"fmt"
	"reflect"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

// Represents a function that updates a poiner to object T by deltaTime.
// You can pass in a pointer-receiving method M for type T with the expression `(*T).M`.
type UpdateFunc[T any] func(object *T, deltaTime float32)

// Represents a function that renders a poiner to object T using the given render context.
// You can pass in a pointer-receiving method M for type T with the expression `(*T).M`.
type RenderFunc[T any] func(object *T, renderContext *render.Context)

// Manages the allocation of a type of game object, reusing memory where possible and issuing object ids.
type Storage[T any] struct {
	data       []T
	owners     []Handle
	active     []bool
	lastActive int           // Index of the last active object. Used to optimize updating.
	UpdateFunc UpdateFunc[T] // Function to call when updating each object.
	RenderFunc RenderFunc[T] // Function to call when rendering each object.
}

var _ StorageOps = (*Storage[any])(nil)

// Creates a new storage that can hold `capacity` number of objects. This capacity will not change, so be generous.
func NewStorage[T any](capacity uint) Storage[T] {
	if capacity == 0 {
		panic("the storage must have capacity greater than 0")
	}
	storage := Storage[T]{
		data:       make([]T, capacity),
		owners:     make([]Handle, capacity),
		active:     make([]bool, capacity),
		lastActive: -1,
	}
	for i := range storage.owners {
		storage.owners[i] = Handle{index: uint16(i), generation: 0, storage: &storage}
	}
	return storage
}

func NewStorageWithFuncs[T any](capacity uint, updateFunc UpdateFunc[T], renderFunc RenderFunc[T]) Storage[T] {
	storage := NewStorage[T](capacity)
	storage.UpdateFunc = updateFunc
	storage.RenderFunc = renderFunc
	return storage
}

// Retrieves a pointer to the object in the storage with the given Id.
// Will return false if the Id is not present or has been overwritten with a different object.
func (st *Storage[T]) Get(h Handle) (*T, bool) {
	if !st.active[h.index] || st.owners[h.index] != h {
		return nil, false
	}
	return &st.data[h.index], true
}

// Retrieves a pointer to the object in the storage with the given Id.
// Will return false if the Id is not present or has been overwritten with a different object.
func (st *Storage[T]) GetUntyped(h Handle) (any, bool) {
	return st.Get(h)
}

// Returns whether the given Id corresponds to an active object in the storage.
func (st *Storage[T]) Has(h Handle) bool {
	return st.active[h.index] && st.owners[h.index] == h
}

// Creates a new entity, returning its Id and a pointer to it. The last result is false if the storage is full.
func (st *Storage[T]) New() (Id[*T], *T, error) {
	for i, active := range st.active {
		if !active {
			if st.owners[i].generation == 0 {
				// Finalize any existing entity that is being overwritten.
				if hasFinalizer, ok := any(&st.data[i]).(engine.HasFinalizer); ok {
					hasFinalizer.Finalize()
				}
			}

			st.active[i] = true
			st.owners[i] = Handle{
				index:      st.owners[i].index,
				generation: st.owners[i].generation + 1,
				storage:    st,
			}

			if hasDefault, ok := any(&st.data[i]).(engine.HasDefault); ok {
				hasDefault.InitDefault()
			} else {
				var empty T
				st.data[i] = empty
			}

			// Update last active index.
			if i >= st.lastActive {
				st.lastActive = i
			}

			return Id[*T]{st.owners[i]}, &st.data[i], nil
		}
	}
	var zero T
	itemType := reflect.TypeOf(zero)
	return Id[*T]{}, nil, fmt.Errorf("ran out of room in storage for %v", itemType.Name())
}

// Marks the object with the given Id as non-active, so that its memory may be reused by a newer object.
// If the Id is already not active, then nothing occurs.
func (st *Storage[T]) Remove(handle Handle) {
	if !st.active[handle.index] || st.owners[handle.index] != handle {
		return
	}
	if hasFinalizer, ok := any(&st.data[handle.index]).(engine.HasFinalizer); ok {
		hasFinalizer.Finalize()
	}
	st.active[handle.index] = false
	if handle.index == uint16(st.lastActive) {
		for i := st.lastActive; i >= 0; i -= 1 {
			if st.active[i] {
				st.lastActive = i
				return
			}
		}
		st.lastActive = -1
	}
}

// Returns an object for iterating through the store.
func (st *Storage[T]) Iter() StorageIter[T] {
	return StorageIter[T]{
		storage: st,
		index:   -1,
	}
}

// Runs an update function on all active objects in the storage.
func (st *Storage[T]) Update(deltaTime float32) {
	if st.UpdateFunc == nil {
		return
	}
	iter := st.Iter()
	for component, _ := iter.Next(); component != nil; component, _ = iter.Next() {
		st.UpdateFunc(component, deltaTime)
	}
}

// Runs a render function on all active objects in the storage.
func (st *Storage[T]) Render(context *render.Context) {
	if st.RenderFunc == nil {
		return
	}
	iter := st.Iter()
	for component, _ := iter.Next(); component != nil; component, _ = iter.Next() {
		st.RenderFunc(component, context)
	}
}

// Deactivates all storage items.
func (st *Storage[T]) Clear() {
	iter := st.Iter()
	for component, handle := iter.Next(); component != nil; component, handle = iter.Next() {
		handle.Remove()
	}
}

func (st *Storage[T]) TearDown() {
	for i := range st.data {
		if hasFinalizer, ok := any(&st.data[i]).(engine.HasFinalizer); ok {
			hasFinalizer.Finalize()
		}
	}
}
