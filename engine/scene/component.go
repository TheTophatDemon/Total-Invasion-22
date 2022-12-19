package scene

type Component interface {
	UpdateComponent(*Scene, Entity, float32)
}

type RenderComponent interface {
	Component
	RenderComponent(*Scene, Entity)
}
