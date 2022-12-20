package scene

type Component interface {
	UpdateComponent(*Scene, Entity, float32)
}

type RenderComponent interface {
	Component
	RenderComponent(*Scene, Entity)
	LayerID() int   // Specifies the order in which this component gets rendered.
	PrepareRender() // Prepares the graphics state to render an instance of this component
}
