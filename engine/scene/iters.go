package scene

import "tophatdemon.com/total-invasion-ii/engine/containers"

func Collect[T any](iter SceneIter[T]) []Handle {
	slice := make([]Handle, 0, iter.Capacity())
	for {
		_, handle := iter.Next()
		if handle.IsNil() {
			break
		}
		slice = append(slice, handle)
	}
	return slice
}

func CollectSet[T any](iter SceneIter[T]) containers.Set[Handle] {
	set := containers.NewSet[Handle](iter.Capacity())
	for {
		_, handle := iter.Next()
		if handle.IsNil() {
			break
		}
		set.Add(handle)
	}
	return set
}
