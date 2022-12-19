package comps

import (
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type MeshRender struct {
	mesh   *assets.Mesh
	shader *assets.Shader
}

func (mr *MeshRender) UpdateComponent(sc *scene.Scene, ent scene.Entity, deltaTime float32) {

}

func (mr *MeshRender) RenderComponent() {

}
