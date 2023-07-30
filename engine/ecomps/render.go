package ecomps

import (
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

// Render default components
func RenderDefaultComps(scene *ecs.Scene, ent ecs.Entity, context *render.Context) {
	meshRender, hasMeshRender := MeshRenders.Get(ent)
	if hasMeshRender {
		meshRender.Render(ent, context)
	}
}
