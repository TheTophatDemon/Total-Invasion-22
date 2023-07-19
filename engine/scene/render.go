package scene

import "github.com/go-gl/mathgl/mgl32"

type RenderComponent interface {
	Component
	RenderComponent(*Scene, Entity, *RenderContext)
	PrepareRender(*RenderContext) // Prepares the graphics state to render an instance of this component
}

// Contains global information useful for rendering.
type RenderContext struct {
	ViewProjection mgl32.Mat4
	View           mgl32.Mat4
	Projection     mgl32.Mat4
	FogStart       float32
	FogLength      float32
}
