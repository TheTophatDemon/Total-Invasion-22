package scene

import (
	"fmt"
	"reflect"
)

type Scene struct {
	entities   []Entity                     // List of active entity IDs, contiguous array.
	active     []bool                       // Indicates which IDs are active, indexed by entity ID.
	components map[reflect.Type][]Component // Stored component slices, keyed by component type, indexed by entity ID.
}

func NewScene() *Scene {
	return &Scene{
		entities:   make([]Entity, 0),
		active:     make([]bool, 0),
		components: make(map[reflect.Type][]Component),
	}
}

func (sc *Scene) Update(deltaTime float32) {
	for _, e := range sc.entities {
		for comp := range sc.components {
			sc.components[comp][e.ID()].UpdateComponent(sc, e, deltaTime)
		}
	}
}

func (sc *Scene) AddEntity() Entity {
	var ent Entity = Entity(ENT_INVALID)
	for i, active := range sc.active {
		if !active {
			ent = Entity(i)
		}
	}
	if ent == ENT_INVALID {
		ent = Entity(len(sc.active))
		sc.active = append(sc.active, true)
	}
	sc.entities = append(sc.entities, ent)
	return ent
}

// Check if entity is in the scene and is active
func (sc *Scene) ValidateEntity(ent Entity) error {
	if len(sc.active) <= int(ent.ID()) || !sc.active[ent.ID()] {
		return fmt.Errorf("could not add components to entity %d because it is not in the scene", ent)
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
		slice, ok := sc.components[cType]

		//Expand the component storage if necessary
		if !ok || len(slice) <= int(ent.ID()) {
			sc.components[cType] = make([]Component, ent.ID()+1)
			if ok {
				copy(sc.components[cType], slice)
			}
		}

		//Assign component
		sc.components[cType][ent.ID()] = c
	}

	return nil
}

// Returns a reference to the given entity's component whose type is the same as the Component passed in.
// Returns an error if the entity isn't active, and returns a nil component if the component isn't found.
func (sc *Scene) GetComponent(ent Entity, componentType Component) (Component, error) {
	if err := sc.ValidateEntity(ent); err != nil {
		return nil, err
	}

	cType := reflect.TypeOf(componentType)

	if len(sc.components[cType]) <= int(ent.ID()) {
		return nil, nil
	}

	ptr := sc.components[cType][ent.ID()]
	return ptr, nil
}

// Returns the given entity's component from the given scene with the type of the component passed in.
// Returns an error if the entity isn't active or doesn't have the given component type.
// Automatically casts the component to the correct concrete type.
func ExtractComponent[C Component](sc *Scene, ent Entity, componentType C) (C, error) {
	zero := *new(C)
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
