package ui

import (
	"runtime"
	"testing"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
)

func initAssets() {
	cache.QuadMesh = geom.CreateMesh(geom.Vertices{
		Pos: []mgl32.Vec3{
			{-1.0, -1.0, 0.0},
			{1.0, -1.0, 0.0},
			{-1.0, 1.0, 0.0},
			{1.0, 1.0, 0.0},
		},
		TexCoord: []mgl32.Vec2{
			{0.0, 1.0},
			{1.0, 1.0},
			{0.0, 0.0},
			{1.0, 0.0},
		},
		Normal: []mgl32.Vec3{
			{0.0, 0.0, 1.0},
			{0.0, 0.0, 1.0},
			{0.0, 0.0, 1.0},
			{0.0, 0.0, 1.0},
		},
		Color: []mgl32.Vec4{
			{1.0, 1.0, 1.0, 1.0},
			{1.0, 1.0, 1.0, 1.0},
			{1.0, 1.0, 1.0, 1.0},
			{1.0, 1.0, 1.0, 1.0},
		},
	}, []uint32{
		1, 2, 0, 1, 3, 2,
	})
}

func checkTranslation(t *testing.T, mat mgl32.Mat4, expected mgl32.Vec3) {
	actual := mat.Col(3).Vec3()
	if !actual.ApproxEqual(expected) {
		_, _, line, _ := runtime.Caller(1)
		t.Logf("line %v: expected translation to be %v, but was %v", line, expected, actual)
		t.FailNow()
	}
}

func TestTransform(t *testing.T) {
	initAssets()
	testPosition := func(t *testing.T, xform Transform, expected mgl32.Vec3) {
		elem := NewBox(xform, nil)
		elem.recalculateMatrices()
		checkTranslation(t, elem.sliceMatrices[textures.SliceCenter], expected)
	}
	t.Run("origin top left anchor top left", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0, 0},
			Anchor:   Ratios{0, 0},
		}, mgl32.Vec3{40, 8})
	})
	t.Run("origin top right anchor top left", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1, 0},
			Anchor:   Ratios{0, 0},
		}, mgl32.Vec3{24, 8})
	})
	t.Run("origin bottom right anchor top left", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1, 1},
			Anchor:   Ratios{0, 0},
		}, mgl32.Vec3{24, -8})
	})
	t.Run("origin bottom left anchor top left", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0, 1},
			Anchor:   Ratios{0, 0},
		}, mgl32.Vec3{40, -8})
	})
	t.Run("origin center anchor top left", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0.5, 0.5},
			Anchor:   Ratios{0, 0},
		}, mgl32.Vec3{32, 0})
	})
	t.Run("origin right center anchor top left", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1.0, 0.5},
			Anchor:   Ratios{0, 0},
		}, mgl32.Vec3{24, 0})
	})
	t.Run("origin top center anchor top left", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0.5, 0.0},
			Anchor:   Ratios{0, 0},
		}, mgl32.Vec3{32, 8})
	})
	// Anchor center
	var screenW, screenH float32
	{
		w, h := engine.ScreenSize()
		screenW = float32(w)
		screenH = float32(h)
	}
	halfScreenW, halfScreenH := float32(screenW/2), float32(screenH/2)
	t.Run("origin top left anchor center", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0, 0},
			Anchor:   Ratios{0.5, 0.5},
		}, mgl32.Vec3{halfScreenW + 40, halfScreenH + 8})
	})
	t.Run("origin top right anchor center", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1, 0},
			Anchor:   Ratios{0.5, 0.5},
		}, mgl32.Vec3{halfScreenW + 24, halfScreenH + 8})
	})
	t.Run("origin bottom right anchor center", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1, 1},
			Anchor:   Ratios{0.5, 0.5},
		}, mgl32.Vec3{halfScreenW + 24, halfScreenH - 8})
	})
	t.Run("origin bottom left anchor center", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0, 1},
			Anchor:   Ratios{0.5, 0.5},
		}, mgl32.Vec3{halfScreenW + 40, halfScreenH - 8})
	})
	t.Run("origin center anchor center", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0.5, 0.5},
			Anchor:   Ratios{0.5, 0.5},
		}, mgl32.Vec3{halfScreenW + 32, halfScreenH})
	})
	t.Run("origin right center anchor center", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1.0, 0.5},
			Anchor:   Ratios{0.5, 0.5},
		}, mgl32.Vec3{halfScreenW + 24, halfScreenH})
	})
	t.Run("origin top center anchor center", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0.5, 0.0},
			Anchor:   Ratios{0.5, 0.5},
		}, mgl32.Vec3{halfScreenW + 32, halfScreenH + 8})
	})

	// Anchor bottom right
	t.Run("origin top left anchor bottom right", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0, 0},
			Anchor:   Ratios{1.0, 1.0},
		}, mgl32.Vec3{screenW + 40, screenH + 8})
	})
	t.Run("origin top right anchor bottom right", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1, 0},
			Anchor:   Ratios{1.0, 1.0},
		}, mgl32.Vec3{screenW + 24, screenH + 8})
	})
	t.Run("origin bottom right anchor bottom right", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1, 1},
			Anchor:   Ratios{1.0, 1.0},
		}, mgl32.Vec3{screenW + 24, screenH - 8})
	})
	t.Run("origin bottom left anchor bottom right", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0, 1},
			Anchor:   Ratios{1.0, 1.0},
		}, mgl32.Vec3{screenW + 40, screenH - 8})
	})
	t.Run("origin center anchor bottom right", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0.5, 0.5},
			Anchor:   Ratios{1.0, 1.0},
		}, mgl32.Vec3{screenW + 32, screenH})
	})
	t.Run("origin right center anchor bottom right", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{1.0, 0.5},
			Anchor:   Ratios{1.0, 1.0},
		}, mgl32.Vec3{screenW + 24, screenH})
	})
	t.Run("origin top center anchor bottom right", func(t *testing.T) {
		testPosition(t, Transform{
			Position: mgl32.Vec2{32},
			Size:     mgl32.Vec2{16, 16},
			Origin:   Ratios{0.5, 0.0},
			Anchor:   Ratios{1.0, 1.0},
		}, mgl32.Vec3{screenW + 32, screenH + 8})
	})
}
