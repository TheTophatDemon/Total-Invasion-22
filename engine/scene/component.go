package scene

type Component interface {
	UpdateComponent(*Scene, Entity, float32)
}
