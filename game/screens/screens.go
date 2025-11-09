package screens

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

type Screen interface {
	Title() string
	Layout(queue *ui.RenderQueue, position mgl32.Vec2, deltaTime float32)
}
