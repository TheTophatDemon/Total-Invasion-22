package scene

import (
	"container/heap"
	"fmt"
	"math"
	"reflect"

	"tophatdemon.com/total-invasion-ii/engine/containers"
)

type Scene struct {
	entities    *containers.List[Entity]     // List of active entity IDs, contiguous array.
	active      []Entity                     // If the entity is active, then it will be in this array indexed by its ID. Otherwise, there will be ENT_INVALID there.
	components  map[reflect.Type][]Component // Stored component slices, keyed by component type, indexed by entity ID.
	renderQueue RenderQueue                  // Sorts entities that are going to be rendered
}

func NewScene() *Scene {
	sc := &Scene{
		entities:    containers.NewList[Entity](),
		active:      make([]Entity, 0),
		components:  make(map[reflect.Type][]Component),
		renderQueue: make(RenderQueue, 0),
	}
	heap.Init(&sc.renderQueue)
	return sc
}

func (sc *Scene) Update(deltaTime float32) {
	for e := sc.entities.Front(); e != nil; e = e.Next() {
		for cType := range sc.components {
			comp := sc.components[cType][e.Value.Index()]
			if comp != nil {
				comp.UpdateComponent(sc, e.Value, deltaTime)
			}
		}
	}
}

func (sc *Scene) Render() {
	lastLayer := math.MinInt
	for _, ri := range sc.renderQueue {
		// Render items should be sorted in the queue so that all items in the same layer are contiguous.
		// Therefore, PrepareRender() needs only be called once when first encountering a given layer.
		if ri.component.LayerID() != lastLayer {
			ri.component.PrepareRender()
		}
		ri.component.RenderComponent(sc, ri.entity)
	}
}

func (sc *Scene) AddEntity() Entity {
	var ent Entity
	for _, active := range sc.active {
		if active == ENT_INVALID {
			ent = active.Renew()
			goto add_to_active
		}
	}
	//If no inactive ID was found, create a new one...
	ent = Entity(len(sc.active))
	sc.active = append(sc.active, ent)
add_to_active:
	sc.entities.PushBack(ent)
	return ent
}

//TODO: Remove entity

// Check if entity is in the scene and is active
func (sc *Scene) ValidateEntity(ent Entity) error {
	if ent == ENT_INVALID || len(sc.active) <= int(ent.Index()) {
		return fmt.Errorf("could not access entity %d because it is not in the scene", ent)
	} else if sc.active[ent.Index()] == ENT_INVALID {
		return fmt.Errorf("could not access entity %d because it is not active", ent)
	}
	return nil
}

func (sc *Scene) AddComponents(ent Entity, components ...Component) error {
	if err := sc.ValidateEntity(ent); err != nil {
		return err
	}

	//Add components
	for _, c := range components {
		cType := reflect.TypeOf(c)
		slice, arrayExists := sc.components[cType]

		//Expand the component storage if necessary
		if !arrayExists || len(slice) <= int(ent.Index()) {
			sc.components[cType] = make([]Component, ent.Index()+1)
			if arrayExists {
				copy(sc.components[cType], slice)
			}
		}

		//Assign component
		sc.components[cType][ent.Index()] = c

		//Assign entity to render queue if it has a render component
		if rc, ok := c.(RenderComponent); ok {
			heap.Push(&sc.renderQueue, &RenderItem{
				component: rc,
				entity:    ent,
			})
		}
	}

	return nil
}

// Returns a reference to the given entity's component whose type is the same as the Component passed in.
// Returns an error if the entity isn't active, and returns a nil component without an error if the component isn't found.
func (sc *Scene) GetComponent(ent Entity, componentType Component) (Component, error) {
	if err := sc.ValidateEntity(ent); err != nil {
		return nil, err
	}

	cType := reflect.TypeOf(componentType)

	if len(sc.components[cType]) <= int(ent.Index()) {
		return nil, nil
	}

	ptr := sc.components[cType][ent.Index()]
	return ptr, nil
}

// Returns the given entity's component from the given scene with the type of the component passed in.
// Returns an error if the entity isn't active or doesn't have the given component type.
// Automatically casts the component to the correct concrete type.
func GetComponent[C Component](sc *Scene, ent Entity, componentType C) (C, error) {
	zero := *new(C) // This should be nil, since C is a pointer to a component
	c, err := sc.GetComponent(ent, componentType)
	if err != nil {
		return zero, err
	}
	typedComponent, ok := c.(C)
	if !ok {
		return zero, fmt.Errorf("component in storage cannot be converted to the appropriate type")
	}
	return typedComponent, nil
}
