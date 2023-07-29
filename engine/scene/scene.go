package scene

import "fmt"

type Scene struct {
	ents   []Entity
	active []bool
}

func NewScene(maxEnts uint) *Scene {
	sc := &Scene{
		ents:   make([]Entity, maxEnts),
		active: make([]bool, maxEnts),
	}
	for i := range sc.ents {
		sc.ents[i] = Entity(i)
	}
	return sc
}

func (scene *Scene) MaxEntCount() uint {
	return uint(len(scene.active))
}

func (sc *Scene) AddEntity() (Entity, error) {
	for i, active := range sc.active {
		if !active {
			newEnt := sc.ents[i].Renew()
			sc.ents[i] = newEnt
			sc.active[i] = true
			return newEnt, nil
		}
	}

	return ENT_INVALID, fmt.Errorf("out of entity indices")
}

func (sc *Scene) RemoveEntity(ent Entity) bool {
	idx := ent.Index()
	if sc.active[idx] {
		sc.active[idx] = false
		return true
	}
	return false
}

type EntsIter struct {
	sc  *Scene
	ent Entity
}

func (ei EntsIter) Valid() bool {
	return ei.sc != nil && ei.ent != ENT_INVALID
}

func (ei EntsIter) Next() EntsIter {
	if !ei.Valid() {
		goto invalid
	}

	for i := int(ei.ent.Index() + 1); i < len(ei.sc.ents); i++ {
		if ei.sc.active[i] {
			return EntsIter{
				sc:  ei.sc,
				ent: ei.sc.ents[i],
			}
		}
	}

invalid:
	return EntsIter{
		sc:  nil,
		ent: ENT_INVALID,
	}
}

func (ei EntsIter) Entity() Entity {
	return ei.ent
}

func (sc *Scene) EntsIter() EntsIter {
	for i := range sc.ents {
		if sc.active[i] {
			return EntsIter{
				sc:  sc,
				ent: sc.ents[i],
			}
		}
	}
	return EntsIter{
		sc:  sc,
		ent: ENT_INVALID,
	}
}
