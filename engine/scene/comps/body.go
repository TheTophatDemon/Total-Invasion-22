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

func (body *Body) Translate(x, y, z float32) {
	body.Position[0] += x
	body.Position[1] += y
	body.Position[2] += z
}

func (body *Body) TranslateV(vec mgl32.Vec3) {
	body.Position = body.Position.Add(vec)
}

func (body *Body) OnLayer(layer collision.Mask) bool {
	return body.Layer.On(layer)
}
