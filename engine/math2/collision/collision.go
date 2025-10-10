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

// When this bit is set, collisions will be bypassed.
const BypassBit Mask = 1 << 63

// Returns true if any of the bits in the provided mask are set on this mask.
// Will return false if otherMask has the bypass bit set.
func (mask Mask) On(otherMask Mask) bool {
	return ((mask|otherMask)&BypassBit == 0) && (mask&otherMask) != 0
}

func (mask Mask) Bypassed() bool {
	return (mask & BypassBit) != 0
}

// Sets the bypass bit on this mask so that it'll temporarily ignore collisions.
func (mask *Mask) SetBypass() {
	if mask != nil {
		*mask |= BypassBit
	}
}

// Unsets the bypass bit so that the mask will respond to collisions again.
func (mask *Mask) ResetBypass() {
	if mask != nil {
		*mask &= (^BypassBit)
	}
}
