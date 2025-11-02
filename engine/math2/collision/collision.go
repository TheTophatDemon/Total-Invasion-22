package collision

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

// Represents the result of a raycast, or collision query.
type Result struct {
	Hit              bool
	Position, Normal mgl32.Vec3
	Distance         float32 // Distance traveled for rays, distance to push out for other objects.
}

func (res Result) String() string {
	return fmt.Sprintf(
		"Result{ Hit: %t, Position: %v, Normal: %v, Distance: %v }",
		res.Hit, res.Position, res.Normal, res.Distance,
	)
}

// Represents a bit mask that filters what things will collide with what.
type Mask uint64

const MaskAll Mask = 0xFFFFFFFFFFFFFFFF

// Returns true if any of the bits in the provided mask are set on this mask.
// Will return false if otherMask has the bypass bit set.
func (mask Mask) On(otherMask Mask) bool {
	return (mask & otherMask) != 0
}
