package scene

type RenderComponent interface {
	Component
	RenderComponent(*Scene, Entity)
	PrepareRender() // Prepares the graphics state to render an instance of this component
}
