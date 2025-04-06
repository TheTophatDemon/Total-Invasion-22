package world

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

var debugShapeCount uint = 0

type DebugShape struct {
	Transform  comps.Transform
	MeshRender comps.MeshRender
	TimeLeft   float32
	id         scene.Id[*DebugShape]
}

func (debugShape *DebugShape) Update(deltaTime float32) {
	if debugShape.TimeLeft <= 0.0 {
		debugShape.id.Remove()
	}
	debugShape.TimeLeft -= deltaTime
}

func (debugShape *DebugShape) Render(context *render.Context) {
	debugShape.MeshRender.Render(&debugShape.Transform, nil, context)
}

func SpawnDebugLine(
	store *scene.Storage[DebugShape],
	start, end mgl32.Vec3,
	duration float32,
	col color.Color,
) (
	id scene.Id[*DebugShape],
	debugShape *DebugShape,
	err error,
) {
	id, debugShape, err = store.New()
	if err != nil {
		return
	}

	debugShape.MeshRender.Mesh = geom.CreateWireMesh(geom.Vertices{
		Pos:   []mgl32.Vec3{start, end},
		Color: []mgl32.Vec4{col.Vector(), col.Vector()},
	}, []uint32{
		0, 1,
	})
	cache.TakeMesh(fmt.Sprintf("!debugshape%v", debugShapeCount), debugShape.MeshRender.Mesh)
	debugShape.MeshRender.Shader = shaders.DebugShader

	debugShape.TimeLeft = duration

	debugShape.id = id

	return
}
