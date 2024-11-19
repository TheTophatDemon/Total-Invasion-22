package collision

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

func TestSphereSweepBoxCollision(t *testing.T) {
	t.Run("direct hit from left side of box", func(t *testing.T) {
		res := SphereSweepBoxCollision(mgl32.Vec3{}, mgl32.Vec3{2.0}, 0.5, math2.BoxFromRadius(0.5).Translate(mgl32.Vec3{2.0}))
		checkRaycastResult(t, res, RaycastResult{
			Hit:      true,
			Distance: 1.0,
			Normal:   mgl32.Vec3{-1.0, 0.0, 0.0},
			Position: mgl32.Vec3{1.5, 0.0, 0.0},
		})
	})
}
