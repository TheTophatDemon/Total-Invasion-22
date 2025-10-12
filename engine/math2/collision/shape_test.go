package collision

import (
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

func TestDistanceFromPoint(t *testing.T) {
	cubeShape := NewBoxShape(0.5, 0.5, 0.5)

	t.Run("distance from side", func(t *testing.T) {
		expected := mgl32.Vec2{0.1, 1.5}.Len()
		dist := cubeShape.DistanceFromPoint(mgl32.Vec3{}, mgl32.Vec3{-0.6, 0.0, -2.0})
		checkDist(t, expected, dist)
	})

	t.Run("distance from top", func(t *testing.T) {
		dist := cubeShape.DistanceFromPoint(mgl32.Vec3{0.0, 1.0, 0.0}, mgl32.Vec3{0.0, 2.0, 0.0})
		checkDist(t, 0.5, dist)
	})

	t.Run("distance from bottom", func(t *testing.T) {
		dist := cubeShape.DistanceFromPoint(mgl32.Vec3{0.0, 1.0, 0.0}, mgl32.Vec3{0.0, -2.0, 0.0})
		checkDist(t, 2.5, dist)
	})

	t.Run("distance from inside", func(t *testing.T) {
		dist := cubeShape.DistanceFromPoint(mgl32.Vec3{3.0, 1.0, 5.0}, mgl32.Vec3{3.1, 0.9, 4.8})
		checkDist(t, 0, dist)
	})
}

func checkDist(t *testing.T, expected, actual float32) {
	if !math2.Almost(actual, expected) {
		t.Fatalf("dist should be %v but was %v", expected, actual)
	}
}
