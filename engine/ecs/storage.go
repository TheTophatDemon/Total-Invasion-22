package ecs

import (
	"fmt"
)

//Generic implementation of component storage
type ComponentStorage[C any] struct {
	storage []C
	owners  []EntID
}

func CreateStorage[C any](initialLength int) *ComponentStorage[C] {
	return &ComponentStorage[C]{
		storage: make([]C, initialLength),
		owners: make([]EntID, initialLength),
	}
}

func (storage *ComponentStorage[C]) Has(id EntID) bool {
	idx, _ := id.Split()
	if idx >= len(storage.owners) { return false }
	return storage.owners[idx] == id
}

func (storage *ComponentStorage[C]) Get(id EntID) (*C, error) {
	idx, _ := id.Split()
	if idx >= len(storage.owners) {
		return nil, fmt.Errorf(ERR_INDEX)
	}
	if storage.owners[idx] != id {
		return nil, fmt.Errorf(ERR_OWNER)
	}
	return &storage.storage[idx], nil
}

func (storage *ComponentStorage[C]) Assign(id EntID, component C) error {
	idx, _ := id.Split()
	if idx >= len(storage.owners) {
		storage.realloc(idx + 1)
	}
	if storage.owners[idx] != NO_ENT && storage.owners[idx] != id {
		return fmt.Errorf(ERR_OWNER)
	}
	storage.storage[idx] = component
	storage.owners[idx] = id
	return nil
}

func (storage *ComponentStorage[C]) Unassign(id EntID) error {
	idx, _ := id.Split()
	if idx >= len(storage.owners) {
		return fmt.Errorf(ERR_INDEX)
	}
	if storage.owners[idx] != id {
		return fmt.Errorf(ERR_OWNER)
	}
	storage.owners[idx] = NO_ENT
	return nil
}

func (storage *ComponentStorage[C]) realloc(newLength int) {
	newStorage := make([]C, newLength)
	newOwners  := make([]EntID, newLength)
	if storage.storage != nil { copy(newStorage, storage.storage) }
	if storage.owners != nil { copy(newOwners, storage.owners) }
	storage.storage = newStorage
	storage.owners = newOwners
}