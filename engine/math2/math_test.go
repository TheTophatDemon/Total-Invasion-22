package math2

import (
	"math"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
)

func TestFrustumFromMatrices(t *testing.T) {
	t.Run("straight ahead", func(t *testing.T) {
		viewProj := mgl32.Perspective(math.Pi/2.0, 1.0, 0.1, 10.0)
		frustum := FrustumFromMatrices(viewProj.Inv())
		if frustum.Planes[PLANE_NEAR].Normal.X() != 0.0 || frustum.Planes[PLANE_NEAR].Normal.Y() != 0.0 || frustum.Planes[PLANE_NEAR].Normal.Z() <= 0.0 {
			t.Errorf("Near plane normal is wrong: %v", frustum.Planes[PLANE_NEAR].Normal)
		}
		if frustum.Planes[PLANE_FAR].Normal.X() != 0.0 || frustum.Planes[PLANE_FAR].Normal.Y() != 0.0 || frustum.Planes[PLANE_FAR].Normal.Z() >= 0.0 {
			t.Errorf("Far plane normal is wrong: %v", frustum.Planes[PLANE_FAR].Normal)
		}
		if frustum.Planes[PLANE_LEFT].Normal.X() >= 0.0 || frustum.Planes[PLANE_LEFT].Normal.Y() != 0.0 || frustum.Planes[PLANE_LEFT].Normal.Z() <= 0.0 {
			t.Errorf("Left plane normal is wrong: %v", frustum.Planes[PLANE_LEFT].Normal)
		}
		if frustum.Planes[PLANE_RIGHT].Normal.X() <= 0.0 || frustum.Planes[PLANE_RIGHT].Normal.Y() != 0.0 || frustum.Planes[PLANE_RIGHT].Normal.Z() <= 0.0 {
			t.Errorf("Right plane normal is wrong: %v", frustum.Planes[PLANE_RIGHT].Normal)
		}
		if frustum.Planes[PLANE_TOP].Normal.X() != 0.0 || frustum.Planes[PLANE_TOP].Normal.Y() <= 0.0 || frustum.Planes[PLANE_TOP].Normal.Z() <= 0.0 {
			t.Errorf("Top plane normal is wrong: %v", frustum.Planes[PLANE_TOP].Normal)
		}
		if frustum.Planes[PLANE_BOTTOM].Normal.X() != 0.0 || frustum.Planes[PLANE_BOTTOM].Normal.Y() >= 0.0 || frustum.Planes[PLANE_BOTTOM].Normal.Z() <= 0.0 {
			t.Errorf("Bottom plane normal is wrong: %v", frustum.Planes[PLANE_BOTTOM].Normal)
		}
	})
}

func TestBoxInFrustum(t *testing.T) {
	viewProj := mgl32.Perspective(math.Pi/2.0, 1.0, 0.1, 10.0)
	invViewProj := viewProj.Inv()
	frustum := FrustumFromMatrices(invViewProj)
	box := BoxFromRadius(1.0).Translate(mgl32.Vec3{0.0, 0.0, -5.0})
	if !frustum.IntersectsBox(box) {
		t.Errorf("Box from %v to %v should intersect frustum", box.Min, box.Max)
	}
	box = box.Translate(mgl32.Vec3{-50.0, 0.0, 0.0})
	if !frustum.IntersectsBox(box) {
		t.Errorf("Box from %v to %v shouldn't intersect the frustum", box.Min, box.Max)
	}
}
