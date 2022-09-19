package engine

const ARENA_INIT_CAPACITY = 16

//Stores and reuses 'things', (basically sync.Pool but type-safe)
type Arena[T any] struct {
	storage []T
	used    []bool
}

//Creates a new arena of things (specified by the type paremeter).
func NewArena[T any]() *Arena[T] {
	return &Arena[T]{
		storage: make([]T, ARENA_INIT_CAPACITY),
		used: make([]bool, ARENA_INIT_CAPACITY),
	}
}

//Retrieves a new or reused thing from the arena.
func ArenaGetNew[T any](arena *Arena[T]) *T {
	for t, thing := range arena.storage {
		if arena.used[t] == false {
			arena.used[t] = true
			return &thing
		}
	}
	//If there is no reusable thing, then add a new thing at the end
	arena.storage = append(arena.storage, *new(T))
	newThing := &arena.storage[len(arena.storage) - 1]
	return newThing
}

//Frees the given element for reuse. Returns false if the element is not in the arena.
func ArenaFree[T any](arena *Arena[T], thing *T) bool {
	for t := range arena.storage {
		if &arena.storage[t] == thing {
			arena.used[t] = false
			return true
		}
	}
	return false
}