package scene

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

// Type-agnostic abstraction of Storage
type StorageOps interface {
	GetUntyped(Handle) (any, bool)
	Has(Handle) bool
	Remove(Handle)
}

// Manages the allocation of a type of game object, reusing memory where possible and issuing object ids.
type Storage[T any] struct {
	data       []T
	owners     []Handle
	active     []bool
	lastActive int // Index of the last active object. Used to optimize updating.
}

var _ StorageOps = (*Storage[any])(nil)

// Represents a function that updates a poiner to object T by deltaTime.
// You can pass in a pointer-receiving method M for type T with the expression `(*T).M`.
type UpdateFunc[T any] func(object *T, deltaTime float32)

// Represents a function that renders a poiner to object T using the given render context.
// You can pass in a pointer-receiving method M for type T with the expression `(*T).M`.
type RenderFunc[T any] func(object *T, renderContext *render.Context)

// Creates a new storage that can hold `capacity` number of objects. This capacity will not change, so be generous.
func NewStorage[T any](capacity uint) *Storage[T] {
	if capacity == 0 {
		panic("the storage must have capacity greater than 0")
	}
	storage := &Storage[T]{
		data:       make([]T, capacity),
		owners:     make([]Handle, capacity),
		active:     make([]bool, capacity),
		lastActive: -1,
	}
	for i := range storage.owners {
		storage.owners[i] = Handle{index: uint16(i), generation: 0, storage: storage}
	}
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
	return Id[*T]{}, nil, fmt.Errorf("no room in storage")
}

// Marks the object with the given Id as non-active, so that its memory may be reused by a newer object.
// If the Id is already not active, then nothing occurs.
func (st *Storage[T]) Remove(h Handle) {
	if !st.active[h.index] || st.owners[h.index] != h {
		return
	}
	st.active[h.index] = false
	if h.index == uint16(st.lastActive) {
		for i := st.lastActive; i >= 0; i -= 1 {
			if st.active[i] {
				st.lastActive = i
				return
			}
		}
		st.lastActive = -1
	}
}

// Runs the given function on all active objects in storage.
func (st *Storage[T]) ForEach(predicate func(*T)) {
	for i := 0; i <= st.lastActive; i += 1 {
		if st.active[i] {
			predicate(&st.data[i])
		}
	}
}

// Runs an update function on all active objects in the storage.
func (st *Storage[T]) Update(updFunc UpdateFunc[T], deltaTime float32) {
	st.ForEach(func(t *T) {
		updFunc(t, deltaTime)
	})
}

// Runs a render function on all active objects in the storage.
func (st *Storage[T]) Render(renderFunc RenderFunc[T], context *render.Context) {
	st.ForEach(func(t *T) {
		renderFunc(t, context)
	})
}

// Returns a closure that returns pointers to each item in storage sequentially, returning nil when the end is reached.
func (st *Storage[T]) Iter() func() (*T, Handle) {
	i := 0
	return func() (*T, Handle) {
		for {
			if i >= len(st.data) || i > st.lastActive {
				return nil, Handle{}
			}
			defer func() { i++ }()
			if st.active[i] {
				return &st.data[i], st.owners[i]
			}
		}
	}
}
