package ecs

import "fmt"

type Scene struct {
	ents   []Entity
	active []bool
}

func NewScene(maxEnts uint) Scene {
	scene := Scene{
		ents:   make([]Entity, maxEnts),
		active: make([]bool, maxEnts),
	}
	for i := range scene.ents {
		scene.ents[i] = Entity(i)
	}
	return scene
}

func (scene *Scene) MaxEntCount() uint {
	return uint(len(scene.active))
}

func (scene *Scene) AddEntity() (Entity, error) {
	for i, active := range scene.active {
		if !active {
			newEnt := scene.ents[i].Renew()
			scene.ents[i] = newEnt
			scene.active[i] = true
			return newEnt, nil
		}
	}

	return ENT_INVALID, fmt.Errorf("out of entity indices")
}

func (scene *Scene) RemoveEntity(ent Entity) bool {
	idx := ent.Index()
	if scene.active[idx] {
		scene.active[idx] = false
		return true
	}
	return false
}

type EntsIter struct {
	scene *Scene
	ent   Entity
}

func (ei EntsIter) Valid() bool {
	return ei.scene != nil && ei.ent != ENT_INVALID
}

func (ei EntsIter) Next() EntsIter {
	if !ei.Valid() {
		goto invalid
	}

	for i := int(ei.ent.Index() + 1); i < len(ei.scene.ents); i++ {
		if ei.scene.active[i] {
			return EntsIter{
				scene: ei.scene,
				ent:   ei.scene.ents[i],
			}
		}
	}

invalid:
	return EntsIter{
		scene: nil,
		ent:   ENT_INVALID,
	}
}

func (ei EntsIter) Entity() Entity {
	return ei.ent
}

func (scene *Scene) EntsIter() EntsIter {
	for i := range scene.ents {
		if scene.active[i] {
			return EntsIter{
				scene: scene,
				ent:   scene.ents[i],
			}
		}
	}
	return EntsIter{
		scene: scene,
		ent:   ENT_INVALID,
	}
}
