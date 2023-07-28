package comps

import (
	"errors"
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type MeshRender struct {
	Mesh   *assets.Mesh
	Shader *assets.Shader
}

func (mr *MeshRender) Render(transforms *scene.ComponentStorage[Transform], ent scene.Entity, context *scene.RenderContext) {
	modelMatrix := mgl32.Ident4()
	transform, ok := transforms.Get(ent)
	if ok {
		modelMatrix = transform.GetMatrix()
	}

	mr.Mesh.Bind()
	mr.Shader.Use()

	err := errors.Join(
		context.SetUniforms(mr.Shader),
		mr.Shader.SetUniformInt(assets.UniformTex, 0),
		mr.Shader.SetUniformInt(assets.UniformAtlas, 1),
		mr.Shader.SetUniformBool(assets.UniformAtlasUsed, false),
		mr.Shader.SetUniformMatrix(assets.UniformModelMatrix, modelMatrix))
	if err != nil {
		fmt.Println(err)
	}

	mr.Mesh.DrawAll()
}
