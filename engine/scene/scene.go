package scene

type Scene struct {
	entities []Entity
}

func NewScene() *Scene {
	return &Scene{
		entities: make([]Entity, 0),
	}
}

func (sc *Scene) Update(deltaTime float32) {
	for _, e := range sc.entities {
		e.Update(deltaTime)
	}
}

func (sc *Scene) AddEntity(ent Entity) {
	sc.entities = append(sc.entities, ent)
}
