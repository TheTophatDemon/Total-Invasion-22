package collision

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

func TestRayPlaneCollision(t *testing.T) {
	t.Run("direct hit", func(t *testing.T) {
		plane := math2.Plane{Normal: mgl32.Vec3{-1.0, 0.0, 0.0}, Dist: 2.0}
		res := RayPlaneCollision(mgl32.Vec3{-5.0, 0.0, 0.0}, mgl32.Vec3{1.0, 0.0, 0.0}, plane, false)
		checkRaycastResult(t, res, RaycastResult{
			Hit:      true,
			Distance: 3.0,
			Normal:   mgl32.Vec3{-1.0, 0.0, 0.0},
			Position: mgl32.Vec3{-2.0, 0.0, 0.0},
		})
	})

	t.Run("don't hit", func(t *testing.T) {
		plane := math2.Plane{Normal: mgl32.Vec3{-1.0, 0.0, 0.0}, Dist: 2.0}
		res := RayPlaneCollision(mgl32.Vec3{+1.0, 0.0, 0.0}, mgl32.Vec3{1.0, 0.0, 0.0}, plane, false)
		checkRaycastResult(t, res, RaycastResult{
			Hit: false,
		})
	})

	t.Run("don't hit from behind", func(t *testing.T) {
		plane := math2.Plane{Normal: mgl32.Vec3{-1.0, 0.0, 0.0}, Dist: 2.0}
		res := RayPlaneCollision(mgl32.Vec3{+1.0, 0.0, 0.0}, mgl32.Vec3{-1.0, 0.0, 0.0}, plane, false)
		checkRaycastResult(t, res, RaycastResult{
			Hit: false,
		})
	})

	t.Run("hit from behind when two sided", func(t *testing.T) {
		plane := math2.Plane{Normal: mgl32.Vec3{-1.0, 0.0, 0.0}, Dist: 2.0}
		res := RayPlaneCollision(mgl32.Vec3{+1.0, 0.0, 0.0}, mgl32.Vec3{-1.0, 0.0, 0.0}, plane, true)
		checkRaycastResult(t, res, RaycastResult{
			Hit:      true,
			Distance: 3.0,
			Normal:   mgl32.Vec3{1.0, 0.0, 0.0},
			Position: mgl32.Vec3{-2.0, 0.0, 0.0},
		})
	})

	t.Run("don't hit when parallel", func(t *testing.T) {
		t.Log("check from front")
		plane := math2.Plane{Normal: mgl32.Vec3{-1.0, 0.0, 0.0}, Dist: 2.0}
		res := RayPlaneCollision(mgl32.Vec3{-5.0, 0.0, 0.0}, mgl32.Vec3{0.0, 1.0, 0.0}, plane, false)
		checkRaycastResult(t, res, RaycastResult{
			Hit: false,
		})
		t.Log("check from behind")
		res = RayPlaneCollision(mgl32.Vec3{0.0, 0.0, 0.0}, mgl32.Vec3{0.0, 1.0, 1.0}.Normalize(), plane, false)
		checkRaycastResult(t, res, RaycastResult{
			Hit: false,
		})
	})

	t.Run("hit diagonally", func(t *testing.T) {
		plane := math2.Plane{Normal: mgl32.Vec3{0.0, 1.0, 0.0}, Dist: 0.0}
		rayOrigin := mgl32.Vec3{-5.0, 5.0, 0.0}
		res := RayPlaneCollision(rayOrigin, mgl32.Vec3{1.0, -1.0, 0.0}.Normalize(), plane, false)
		checkRaycastResult(t, res, RaycastResult{
			Hit:      true,
			Distance: rayOrigin.Len(),
			Normal:   mgl32.Vec3{0.0, 1.0, 0.0},
			Position: mgl32.Vec3{0.0, 0.0, 0.0},
		})
	})
}

func checkRaycastResult(t *testing.T, actual, expected RaycastResult) {
	if actual.Hit != expected.Hit {
		if expected.Hit {
			t.Error("ray should have hit")
		} else {
			t.Error("ray should not have hit")
		}
	}
	if actual.Distance != expected.Distance {
		t.Errorf("distance should be %v but is %v", expected.Distance, actual.Distance)
	}
	if !actual.Normal.ApproxEqual(expected.Normal) {
		t.Errorf("normal should be %v but is %v", expected.Normal, actual.Normal)
	}
	if !actual.Position.ApproxEqual(expected.Position) {
		t.Errorf("position should be %v but is %v", expected.Position, actual.Position)
	}
}
