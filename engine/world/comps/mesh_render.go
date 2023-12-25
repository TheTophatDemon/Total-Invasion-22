package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type MeshRender struct {
	Mesh    *geom.Mesh
	Shader  *shaders.Shader
	Texture *textures.Texture
	Group   string
}

func NewMeshRender(Mesh *geom.Mesh, Shader *shaders.Shader, Texture *textures.Texture) MeshRender {
	return NewMeshRenderGroup(Mesh, Shader, Texture, "")
}

func NewMeshRenderGroup(Mesh *geom.Mesh, Shader *shaders.Shader, Texture *textures.Texture, Group string) MeshRender {
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
	if mr.Mesh == nil || mr.Shader == nil {
		return
	}

	var modelMatrix mgl32.Mat4
	if transform != nil {
		modelMatrix = transform.Matrix()
	} else {
		modelMatrix = mgl32.Ident4()
	}

	// Bind resources
	mr.Mesh.Bind()
	mr.Shader.Use()
	if mr.Texture != nil {
		mr.Texture.Bind()
	}

	// Set uniforms
	_ = context.SetUniforms(mr.Shader)
	_ = mr.Shader.SetUniformInt(shaders.UniformTex, 0)
	_ = mr.Shader.SetUniformMatrix(shaders.UniformModelMatrix, modelMatrix)
	if animPlayer != nil {
		_ = mr.Shader.SetUniformVec4(shaders.UniformSrcRect, animPlayer.FrameUV().Vec4())
	} else {
		_ = mr.Shader.SetUniformVec4(shaders.UniformSrcRect, mgl32.Vec4{0.0, 0.0, 1.0, 1.0})
	}

	if len(mr.Group) == 0 {
		mr.Mesh.DrawAll()
	} else {
		mr.Mesh.DrawGroup(mr.Group)
	}
}
