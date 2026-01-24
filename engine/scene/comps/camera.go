package comps

type Camera struct {
	Transform
}

func NewCamera(transform Transform) Camera {
	return Camera{
		Transform: transform,
	}
}
