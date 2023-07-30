package ecs

type ComponentStorage[C any] struct {
	components []C
	user       []Entity // Stores the entity ID using the component. Will be set to ENT_INVALID if the component is unused.
}

func RegisterComponent[C any](scene *Scene) *ComponentStorage[C] {
	storage := &ComponentStorage[C]{
		components: make([]C, scene.MaxEntCount()),
		user:       make([]Entity, scene.MaxEntCount()),
	}
	for u := range storage.user {
		storage.user[u] = ENT_INVALID
	}
	return storage
}

func (cs *ComponentStorage[C]) Get(entity Entity) (*C, bool) {
	idx := entity.Index()
	if cs.user[idx] != entity {
		return nil, false
	}
	return &cs.components[idx], true
}

func (cs *ComponentStorage[C]) Has(entity Entity) bool {
	idx := entity.Index()
	return cs.user[idx] == entity
}

func (cs *ComponentStorage[C]) Assign(entity Entity, init C) {
	idx := entity.Index()
	cs.user[idx] = entity
	cs.components[idx] = init
}

func (cs *ComponentStorage[C]) Remove(entity Entity) {
	idx := entity.Index()
	cs.user[idx] = ENT_INVALID
}
