package ecs

//Entity record for bookkeeping
type rEnt struct {
	id        EntID
	dead      bool
}

type Predicate func(EntID)bool

// type Group struct {
// 	ents      []EntID
// 	predicate Predicate
// }

type World struct {
	ents   []rEnt
	// groups []Group
}

//Returns the "indexing" part of the ID and the "generation" part as separate values
func (id EntID) Split() (int, int) {
	return int(id & ENT_IDX_MASK), int(id & ENT_GEN_MASK)
}

func CreateWorld(capacity int) *World {
	world := &World {
		ents: make([]rEnt, 1, capacity),
	}
	world.ents[0].dead = true //Entity 0 is used as a placeholder and must never be used
	return world
}

func (w *World) NewEnt() EntID {
	for e, ent := range w.ents {
		if e != int(NO_ENT) && ent.dead {
			//Reuse the dead entity, incrementing its generation count
			newID := (ent.id & ENT_IDX_MASK) | ((ent.id & ENT_GEN_MASK) + GEN_INC)
			w.ents[e] = rEnt{ id: newID }
			return newID
		}
	}
	//If no reusable entities, expand the list.
	newID := EntID((len(w.ents)) & ENT_IDX_MASK)
	w.ents = append(w.ents, rEnt { id: newID })
	return newID
}

func (w *World) KillEnt(id EntID) bool {
	idx, gen := id.Split()
	if idx >= len(w.ents) { return false } //Don't remove if out of bounds
	if _, gen2 := w.ents[idx].id.Split(); gen2 > gen { return false } //Don't remove if entity was already replaced
	//Mark the entity as dead so it may be reused.
	w.ents[idx].dead = true
	return true
}

//Returns true if the entity with the given ID is no longer active
func (w *World) IsDead(id EntID) bool {
	idx, gen := id.Split()
	_, gen2 := w.ents[idx].id.Split()
	return (gen2 > gen) || (w.ents[idx].dead)
}

func (w *World) Query(predicate Predicate) []EntID {
	result := make([]EntID, 0)
	for _, ent := range w.ents {
		if !ent.dead && predicate(ent.id) {
			result = append(result, ent.id)
		}
	}
	return result
}