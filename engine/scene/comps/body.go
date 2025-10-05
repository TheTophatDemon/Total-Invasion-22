package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
)

// Represents a game entity that moves and/or stops other things from moving.
type Body struct {
	Position, Velocity mgl32.Vec3
	Shape              collision.Shape
	Layer              collision.Mask // The collision layer(s) that this body resides on
}

func (body *Body) Body() *Body {
	return body
}

func (body *Body) OnLayer(layer collision.Mask) bool {
	return body.Layer.On(layer)
}
