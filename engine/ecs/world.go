package ecs

//First 16 bits are the actual ID, the last 16 bits are the generation count.
//The generation count increases with each reuse of the ID to indicate that it's a different entity than the last one.
type EntID uint32

const (
	ENT_IDX_MASK  = 0x0000FFFF
	ENT_GEN_MASK  = 0xFFFF0000
	GEN_INC       = 0x00010000 //Number to add to increase the generation by one (relative to the last 2 bytes)
)

const (
	ERR_INDEX =    "Entity index is out of bounds."
	ERR_OWNER =    "Entity does not own the component, or was overwritten."
	ERR_CTYPE =    "Given component is of the wrong type."
	ERR_REGISTER = "Component type was not registered."
)

//Entity record for bookkeeping
type rEnt struct {
	id        EntID
	dead      bool //Marks the entity ID for potential reuse
}

type Predicate func(EntID)bool //A filter function that determines an entity's group membership

type Group struct {
	ents      []EntID //Cache of ent Ids that satisfy the group's predicate
	predicate Predicate
}

type World struct {
	ents   []rEnt
	groups []Group
}

//Returns the "indexing" part of the ID and the "generation" part as separate values
func (id EntID) Split() (int, int) {
	return int(id & ENT_IDX_MASK), int(id & ENT_GEN_MASK)
}

func CreateWorld(capacity int) *World {
	world := &World {
		ents: make([]rEnt, 0, capacity),
	}
	return world
}

func (w *World) NewEnt() EntID {
	for e, ent := range w.ents {
		if ent.dead {
			//Reuse the dead entity, incrementing its generation count
			newID := (ent.id & ENT_IDX_MASK) | ((ent.id & ENT_GEN_MASK) + GEN_INC)
			w.ents[e] = rEnt{ id: newID, dead: false }
			return newID
		}
	}
	//If no reusable entities, expand the list.
	newID := EntID((len(w.ents)) & ENT_IDX_MASK)
	w.ents = append(w.ents, rEnt { id: newID, dead: false })
	return newID
}

//Marks an entity as dead so that its resources may be reused.
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

//Adds an entity group to the world that filters entities based on the given function.
// func (w *World) AddGroup(pred Predicate) *Group {
// 	w.groups = append(w.groups, Group{
// 		ents: make([]EntID, 0),
// 		predicate: pred,
// 	})

// 	group := &w.groups[len(w.groups)]
// 	//Scan for existing ents that belong in the group
// 	for _, rent := range w.ents {
// 		if !rent.dead && pred(rent.id) {
// 			group.ents = append(group.ents, rent.id)
// 		}
// 	}
// 	return group
// }

func (w *World) Query(predicate Predicate) []EntID {
	result := make([]EntID, 0)
	for _, ent := range w.ents {
		if !ent.dead && predicate(ent.id) {
			result = append(result, ent.id)
		}
	}
	return result
}