package ecs

import (
	"fmt"
)

const STORAGE_INIT_CAPACITY = 16
const UNOWNED_IDX = 0xFFFFFFFF

//Generic implementation of component storage
type ComponentStorage[C any] struct {
	components    []C     //Contiguous array of components
	owners        []EntID //Indicates the Entity ID that owns each component
	sparseIndexes []int   //Using the Entity ID as an index, this array points to the component in storage that it owns.
}

func CreateStorage[C any]() *ComponentStorage[C] {
	storage := &ComponentStorage[C]{
		components: make([]C, 0, STORAGE_INIT_CAPACITY),
		owners: make([]EntID, 0, STORAGE_INIT_CAPACITY),
		sparseIndexes: make([]int, 0),
	}
	return storage
}

func (storage *ComponentStorage[C]) Has(id EntID) bool {
	idx, _ := id.Split()
	if idx >= len(storage.sparseIndexes) { return false }
	return storage.sparseIndexes[idx] != UNOWNED_IDX
}

func (storage *ComponentStorage[C]) Get(id EntID) (*C, error) {
	idx, _ := id.Split()
	if idx >= len(storage.sparseIndexes) {
		return nil, fmt.Errorf(ERR_INDEX)
	}
	ci := storage.sparseIndexes[idx]
	if ci == UNOWNED_IDX {
		return nil, fmt.Errorf(ERR_OWNER)
	} else if ci >= len(storage.components) {
		return nil, fmt.Errorf("Sparse array index is out of bounds.")
	}
	return &storage.components[ci], nil
}

// func (storage *ComponentStorage[C]) Iter() func()(bool, *C) {
// 	n := 0
// 	return func()(bool, *C) {
// 		if n < len(storage.components) {
// 			n += 1
// 			return true, &storage.components[n - 1]
// 		} else {
// 			return false, nil
// 		}
// 	}
// }

func (storage *ComponentStorage[C]) ForEach(fn func(*C)) {
	for c := range storage.components {
		fn(&storage.components[c])
	}
}

func (storage *ComponentStorage[C]) Assign(id EntID, comp C) (*C, error) {
	idx, _ := id.Split()
	//Expand the sparse index array if necessary
	if idx >= len(storage.sparseIndexes) {
		//Reallocate sparse indexes array to include the new ID
		newIndexes := make([]int, idx + 1)
		copy(newIndexes, storage.sparseIndexes)
		//Indicate unused indexes
		for i := len(storage.sparseIndexes); i < len(newIndexes); i++ {
			newIndexes[i] = UNOWNED_IDX
		}
		storage.sparseIndexes = newIndexes
	}
	//Return component if previously assigned.
	if ci := storage.sparseIndexes[idx]; ci != UNOWNED_IDX {
		return &storage.components[ci], nil
	}
	//Otherwise look for free component
	compIndex := len(storage.components)
	for o, owner := range storage.owners {
		if owner == UNOWNED_IDX {
			compIndex = o
		}
	}
	if compIndex == len(storage.components) {
		//Expand the components/owners arrays as necessary.
		storage.components = append(storage.components, *new(C))
		storage.owners = append(storage.owners, UNOWNED_IDX)
	}
	storage.sparseIndexes[idx] = compIndex
	storage.owners[compIndex] = id
	storage.components[compIndex] = comp
	return &storage.components[compIndex], nil
}

//Mark the component for this entity as having no owner, so it can be reused by other entities.
func (storage *ComponentStorage[C]) Unassign(id EntID) error {
	idx, _ := id.Split()
	if idx >= len(storage.sparseIndexes) {
		return fmt.Errorf(ERR_INDEX)
	}
	storage.owners[storage.sparseIndexes[idx]] = UNOWNED_IDX
	storage.sparseIndexes[idx] = UNOWNED_IDX
	return nil
}