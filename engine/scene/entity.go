package scene

import "reflect"

type Entity struct {
	components map[reflect.Type]Component
}

// Creates a new entity with the given components.
func NewEntity(components ...Component) Entity {
	// Initialize entity
	ent := Entity{
		components: make(map[reflect.Type]Component, 8),
	}
	// Add components from the parameters
	for _, c := range components {
		ent.components[reflect.TypeOf(c)] = c
	}

	return ent
}

// If the entity has a component of the given type, then the passed pointer is updated to point to it.
// Returns true if the entity was found, and false otherwise.
func (ent *Entity) GetComponent(cType Component) Component {
	comp, ok := ent.components[reflect.TypeOf(cType)]
	if !ok {
		return nil
	} else {
		return comp
	}
}

// Adds the given component to the entity, unless that component is already added.
// The function returns false if the component was already added, or if its type is invalid.
func (ent *Entity) AddComponent(component Component) bool {
	cType := reflect.TypeOf(component)
	_, exists := ent.components[cType]
	if !exists {
		ent.components[cType] = component
		return true
	}
	return false
}

// Removes the given component from the entity.
// Returns true if removal was successful, or false if the entity didn't have the given component.
func (ent *Entity) RemoveComponent(component Component) bool {
	cType := reflect.TypeOf(component)
	_, exists := ent.components[cType]
	if exists {
		delete(ent.components, cType)
		return true
	}
	return false
}

// Updates all of the components within the entity.
func (ent *Entity) Update(deltaTime float32) {
	for compName := range ent.components {
		ent.components[compName].Update(ent, deltaTime)
	}
}
