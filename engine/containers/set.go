package containers

import (
	"maps"
	"slices"
)

type Set[Key comparable] map[Key]struct{}

func NewSet[Key comparable](capacity int) Set[Key] {
	return make(Set[Key], capacity)
}

// Adds an item to the set. Replaces existing item in that position. Returns true if an existing item was replaced.
func (set Set[Key]) Add(item Key) {
	set[item] = struct{}{}
}

func (set Set[Key]) Collect() []Key {
	return slices.Collect(maps.Keys(set))
}
