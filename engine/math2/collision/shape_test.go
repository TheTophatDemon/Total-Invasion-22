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

	t.Run("NaN is infinite distance", func(t *testing.T) {
		dist := cubeShape.DistanceFromPoint(mgl32.Vec3{math2.NaN[float32](), 1.0, 5.0}, mgl32.Vec3{3.1, 0.9, 4.8})
		if !math2.IsInf(dist, 0) {
			t.Fatalf("dist should be infinite")
		}
		dist = cubeShape.DistanceFromPoint(mgl32.Vec3{0.0, 1.0, 5.0}, mgl32.Vec3{3.1, math2.NaN[float32](), 4.8})
		if !math2.IsInf(dist, 0) {
			t.Fatalf("dist should be infinite")
		}
	})
}

func checkDist(t *testing.T, expected, actual float32) {
	if !math2.Almost(actual, expected) {
		t.Fatalf("dist should be %v but was %v", expected, actual)
	}
}

func TestRaycast(t *testing.T) {
	cubeShape := NewBoxShape(0.5, 0.5, 0.5)

	t.Run("cast from top", func(t *testing.T) {
		result := cubeShape.Raycast(
			mgl32.Vec3{1.0, 2.0, 3.0},
			mgl32.Vec3{1.0, 3.0, 3.0},
			mgl32.Vec3{0.0, -1.0, 0.0},
			10.0,
		)
		checkResult(t, Result{
			Hit:      true,
			Position: mgl32.Vec3{1.0, 2.5, 3.0},
			Normal:   mgl32.Vec3{0.0, 1.0, 0.0},
			Distance: 0.5,
		}, result)
	})

	t.Run("cast from top but it misses", func(t *testing.T) {
		result := cubeShape.Raycast(
			mgl32.Vec3{1.0, 2.0, 3.0},
			mgl32.Vec3{1.0, 3.0, 3.0},
			mgl32.Vec3{0.0, -1.0, 0.0},
			0.25,
		)
		checkResult(t, Result{
			Hit:      false,
			Position: mgl32.Vec3{},
			Normal:   mgl32.Vec3{},
			Distance: 0.25,
		}, result)
	})

	t.Run("cast from bottom", func(t *testing.T) {
		result := cubeShape.Raycast(
			mgl32.Vec3{1.0, 2.0, 3.0},
			mgl32.Vec3{1.0, 1.0, 3.0},
			mgl32.Vec3{0.0, 1.0, 0.0},
			10.0,
		)
		checkResult(t, Result{
			Hit:      true,
			Position: mgl32.Vec3{1.0, 1.5, 3.0},
			Normal:   mgl32.Vec3{0.0, -1.0, 0.0},
			Distance: 0.5,
		}, result)
	})

	t.Run("cast from bottom but it misses", func(t *testing.T) {
		result := cubeShape.Raycast(
			mgl32.Vec3{1.0, 2.0, 3.0},
			mgl32.Vec3{1.0, 1.0, 3.0},
			mgl32.Vec3{0.0, 1.0, 0.0},
			0.25,
		)
		checkResult(t, Result{
			Hit:      false,
			Position: mgl32.Vec3{},
			Normal:   mgl32.Vec3{},
			Distance: 0.25,
		}, result)
	})

	t.Run("cast from side", func(t *testing.T) {
		result := cubeShape.Raycast(
			mgl32.Vec3{1.0, 2.0, 3.0},
			mgl32.Vec3{0.0, 2.0, 3.0},
			mgl32.Vec3{1.0, 0.0, 0.0},
			10.0,
		)
		checkResult(t, Result{
			Hit:      true,
			Position: mgl32.Vec3{0.5, 2.0, 3.0},
			Normal:   mgl32.Vec3{-1.0, 0.0, 0.0},
			Distance: 0.5,
		}, result)
	})

	t.Run("cast from side but it misses", func(t *testing.T) {
		result := cubeShape.Raycast(
			mgl32.Vec3{1.0, 2.0, 3.0},
			mgl32.Vec3{0.0, 2.0, 3.0},
			mgl32.Vec3{1.0, 0.0, 0.0},
			0.25,
		)
		checkResult(t, Result{
			Hit:      false,
			Position: mgl32.Vec3{},
			Normal:   mgl32.Vec3{},
			Distance: 0.25,
		}, result)
	})
}

func checkResult(t *testing.T, expected, actual Result) {
	fail := expected.Hit != actual.Hit ||
		!expected.Position.ApproxEqual(actual.Position) ||
		!expected.Normal.ApproxEqual(actual.Normal) ||
		!math2.Almost(expected.Distance, actual.Distance)
	if fail {
		t.Fatalf("result should be %v but was %v", expected, actual)
	}
}
