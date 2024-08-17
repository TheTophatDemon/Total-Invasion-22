package cache

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
)

var (
	// A quad on the XY plane centered at (0,0) with a width and height of 2.
	QuadMesh *geom.Mesh
)

// Initialize built-in assets
func InitBuiltInAssets() {
	shaders.Init()

	QuadMesh = geom.CreateMesh(geom.Vertices{
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

func FreeBuiltInAssets() {
	shaders.Free()
}
