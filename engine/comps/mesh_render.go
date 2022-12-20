package comps

import (
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

var meshRenderLayers map[MeshRender]int
var lastUnusedLayer int = 0

func init() {
	meshRenderLayers = make(map[MeshRender]int)
}

type MeshRender struct {
	mesh   *assets.Mesh
	shader *assets.Shader
}

func (mr *MeshRender) UpdateComponent(sc *scene.Scene, ent scene.Entity, deltaTime float32) {

}

func (mr *MeshRender) RenderComponent(sc *scene.Scene, ent scene.Entity) {
	mr.mesh.DrawAll()
}

func (mr *MeshRender) PrepareRender() {
	mr.mesh.Bind()
	mr.shader.Use()

	// TODO: Set shader uniforms
	// TODO: Bind the texture
}

func (mr *MeshRender) LayerID() int {
	// TODO: Cache the layer ID...somehow
	layer, ok := meshRenderLayers[*mr]
	if !ok {
		layer = lastUnusedLayer
		lastUnusedLayer += 1
		meshRenderLayers[*mr] = layer
	}
	return layer
}
