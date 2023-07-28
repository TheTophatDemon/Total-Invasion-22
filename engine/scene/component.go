package scene

type ComponentStorage[C any] struct {
	components []C
	used       []bool
}

func RegisterComponent[C any](scene *Scene) *ComponentStorage[C] {
	return &ComponentStorage[C]{
		components: make([]C, scene.MaxEntCount()),
		used:       make([]bool, scene.MaxEntCount()),
	}
}

func (cs *ComponentStorage[C]) Get(entity Entity) (*C, bool) {
	idx := entity.Index()
	if cs.used[idx] == false {
		return nil, false
	}
	return &cs.components[idx], true
}

func (cs *ComponentStorage[C]) Has(entity Entity) bool {
	idx := entity.Index()
	return cs.used[idx]
}

func (cs *ComponentStorage[C]) Assign(entity Entity, init C) {
	idx := entity.Index()
	cs.used[idx] = true
	cs.components[idx] = init
}

func (cs *ComponentStorage[C]) Remove(entity Entity) {
	idx := entity.Index()
	cs.used[idx] = false
}
