package comps

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type MeshRender struct {
	Mesh    *assets.Mesh
	Shader  *assets.Shader
	Texture *assets.Texture
	Group   string
}

func NewMeshRender(Mesh *assets.Mesh, Shader *assets.Shader, Texture *assets.Texture) MeshRender {
	return NewMeshRenderGroup(Mesh, Shader, Texture, "")
}

func NewMeshRenderGroup(Mesh *assets.Mesh, Shader *assets.Shader, Texture *assets.Texture, Group string) MeshRender {
	return MeshRender{
		Mesh,
		Shader,
		Texture,
		Group,
	}
}

// Renders the mesh with the given local transform and the optional animation player.
// If transform is nil, then the model matrix is the identity matrix.
func (mr *MeshRender) Render(
	transform *Transform,
	animPlayer *AnimationPlayer,
	context *render.Context,
) {
	var modelMatrix mgl32.Mat4
	if transform != nil {
		modelMatrix = transform.GetMatrix()
	} else {
		modelMatrix = mgl32.Ident4()
	}

	// Bind resources
	if mr.Mesh == nil || mr.Shader == nil {
		fmt.Println("WARNING: MeshRender is missing mesh or shader.")
	}
	mr.Mesh.Bind()
	mr.Shader.Use()
	if mr.Texture != nil {
		mr.Texture.Bind()
	}

	frame := 0
	if animPlayer != nil {
		frame = animPlayer.Frame()
	}

	// Set uniforms
	_ = context.SetUniforms(mr.Shader)
	_ = mr.Shader.SetUniformInt(assets.UniformTex, 0)
	_ = mr.Shader.SetUniformInt(assets.UniformAtlas, 1)
	_ = mr.Shader.SetUniformBool(assets.UniformAtlasUsed, mr.Texture.IsAtlas())
	_ = mr.Shader.SetUniformMatrix(assets.UniformModelMatrix, modelMatrix)
	_ = mr.Shader.SetUniformInt(assets.UniformFrame, frame)

	if len(mr.Group) == 0 {
		mr.Mesh.DrawAll()
	} else {
		mr.Mesh.DrawGroup(mr.Group)
	}
}
