package ecomps

import (
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

// Render default components
func RenderDefaultComps(sc *scene.Scene, ent scene.Entity, context *scene.RenderContext) {
	meshRender, hasMeshRender := MeshRenderComps.Get(ent)
	if hasMeshRender {
		meshRender.Render(ent, context)
	}
}
