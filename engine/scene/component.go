package scene

type Component interface {
	Update(*Entity, float32)
}
