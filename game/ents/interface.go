package ents

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

// In order to prevent a circular dependency of packages, entities interact with the World through this interface.
// As a bonus, this prevents entities from doing things with the world that they shouldn't, like running a full update.
type WorldOps interface {
	ShowMessage(text string, duration float32, priority int, colr color.Color)
	Raycast(rayOrigin, rayDir mgl32.Vec3, includeBodies bool, maxDist float32, excludeBody *comps.Body) (collision.RaycastResult, comps.HasBody)
}
