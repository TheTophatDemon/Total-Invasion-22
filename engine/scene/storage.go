package scene

import (
	"tophatdemon.com/total-invasion-ii/engine/render"
)

// TODO: Have the handle be a struct instead..? Would have to make a Storage interface.
type Handle interface {
	GetUntyped() (any, bool)
	Exists() bool
	Remove()
	Equals(Handle) bool
	Index() uint16
	Generation() uint16
}

type Id[T any] struct {
	index, generation uint16
	storage           *Storage[T]
}

var _ Handle = (*Id[int])(nil)

func GetTyped[T any](handle Handle) (T, bool) {
	var empty T
	data, exists := handle.GetUntyped()
	if !exists {
		return empty, false
	}
	typedData, isType := data.(T)
	if !isType {
		return empty, false
	}
	return typedData, true
}

func (id Id[T]) Index() uint16 {
	return id.index
}

func (id Id[T]) Generation() uint16 {
	return id.generation
}

func (id Id[T]) Get() (*T, bool) {
	if id.storage == nil {
		return nil, false
	}
	ptr, ok := id.storage.Get(id)
	return ptr, ok
}

func (id Id[T]) Equals(handle Handle) bool {
	return id.index == handle.Index() && id.generation == handle.Generation()
}

func (id Id[T]) GetUntyped() (any, bool) {
	return id.Get()
}

func (id Id[T]) Exists() bool {
	if id.storage == nil {
		return false
	}
	return id.storage.Has(id)
}

func (id Id[T]) Remove() {
	if id.storage == nil {
		return
	}
	id.storage.Remove(id)
}

func (id Id[T]) IsNil() bool {
	return id.storage == nil
}

// Manages the allocation of a type of game object, reusing memory where possible and issuing object ids.
type Storage[T any] struct {
	data       []T
	owners     []Id[T]
	active     []bool
	lastActive int // Index of the last active object. Used to optimize updating.
}

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
		owners:     make([]Id[T], capacity),
		active:     make([]bool, capacity),
		lastActive: -1,
	}
	for i := range storage.owners {
		storage.owners[i] = Id[T]{index: uint16(i), generation: 0}
	}
	return storage
}

// Retrieves a pointer to the object in the storage with the given Id.
// Will return false if the Id is not present or has been overwritten with a different object.
func (st *Storage[T]) Get(id Id[T]) (*T, bool) {
	if !st.active[id.index] || st.owners[id.index] != id {
		return nil, false
	}
	return &st.data[id.index], true
}

// Returns whether the given Id corresponds to an active object in the storage.
func (st *Storage[T]) Has(id Id[T]) bool {
	return st.active[id.index] && st.owners[id.index] == id
}

// Creates a new entity, returning its Id and a pointer to it. The last result is false if the storage is full.
func (st *Storage[T]) New(init T) (Id[T], *T, bool) {
	for i, active := range st.active {
		if !active {
			st.active[i] = true
			st.owners[i] = Id[T]{
				index:      st.owners[i].index,
				generation: st.owners[i].generation + 1,
				storage:    st,
			}
			st.data[i] = init

			// Update last active index.
			if i >= st.lastActive {
				st.lastActive = i
			}

			return st.owners[i], &st.data[i], true
		}
	}
	return Id[T]{index: 0, generation: 0, storage: nil}, nil, false
}

// Marks the object with the given Id as non-active, so that its memory may be reused by a newer object.
// If the Id is already not active, then nothing occurs.
func (st *Storage[T]) Remove(id Id[T]) {
	if !st.active[id.index] || st.owners[id.index] != id {
		return
	}
	st.active[id.index] = false
	if id.index == uint16(st.lastActive) {
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
func (st *Storage[T]) Iter() func() (*T, Id[T]) {
	i := 0
	return func() (*T, Id[T]) {
		for {
			if i >= len(st.data) || i > st.lastActive {
				return nil, Id[T]{}
			}
			defer func() { i++ }()
			if st.active[i] {
				return &st.data[i], st.owners[i]
			}
		}
	}
}
