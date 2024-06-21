package collision

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

func TestRaySphereCollision(t *testing.T) {
	t.Run("simple example", func(t *testing.T) {
		sphere := NewSphere(1.0)
		res := sphere.Raycast(mgl32.Vec3{}, mgl32.Vec3{1.0}, mgl32.Vec3{10.0, 0.0, 0.0}, 100.0)
		if !res.Hit {
			t.Errorf("raycast did not hit")
		}
		if math2.Abs(res.Distance-9.0) > 0.01 {
			t.Errorf("raycast distance was not 9, it was %v", res.Distance)
		}
		if !res.Normal.ApproxEqual(mgl32.Vec3{-1.0, 0.0, 0.0}) {
			t.Errorf("raycast normal did not point along -X. Instead it was (%v, %v, %v)", res.Normal[0], res.Normal[1], res.Normal[2])
		}
		if !res.Position.ApproxEqual(mgl32.Vec3{9.0, 0.0, 0.0}) {
			t.Errorf("raycast hit position was not at the edge of the sphere. Instead it was (%v, %v, %v)", res.Position[0], res.Position[1], res.Position[2])
		}
	})
}
