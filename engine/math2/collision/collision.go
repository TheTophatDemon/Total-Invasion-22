package collision

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Represents the result of a raycast, or collision query.
type Result struct {
	Hit              bool
	Position, Normal mgl32.Vec3
	Distance         float32 // Distance traveled for rays, distance to push out for other objects.
}

// Represents a bit mask that filters what things will collide with what.
type Mask uint64
