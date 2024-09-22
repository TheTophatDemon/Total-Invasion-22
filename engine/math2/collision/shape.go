package collision

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Shape interface {
	String() string
	Extents() math2.Box                                                               // Returns the body's bounding box, centered at its origin.
	Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) RaycastResult // Test the shape for collision against a ray
	ResolveCollision(myPosition, theirPosition mgl32.Vec3, theirShape Shape) Result
}

type shape struct {
	extents math2.Box
}
