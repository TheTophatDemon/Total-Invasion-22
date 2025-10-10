package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type MeshRender struct {
	Mesh           *geom.Mesh
	Shader         *shaders.Shader
	Texture        *textures.Texture
	LocalTransform Transform // Transform relative to the rendered position
	Group          string
}

func NewMeshRender(Mesh *geom.Mesh, Shader *shaders.Shader, Texture *textures.Texture) MeshRender {
	return NewMeshRenderGroup(Mesh, Shader, Texture, "")
}

func NewMeshRenderGroup(mesh *geom.Mesh, shader *shaders.Shader, texture *textures.Texture, group string) MeshRender {
	return MeshRender{
		Mesh:    mesh,
		Shader:  shader,
		Texture: texture,
		Group:   group,
	}
}

// Renders the mesh with the given local transform and the optional animation player.
func (mr *MeshRender) Render(
	position mgl32.Vec3,
	animPlayer *AnimationPlayer,
	context *render.Context,
) {
	if mr == nil || mr.Mesh == nil || mr.Shader == nil {
		return
	}

	modelMatrix := mgl32.Translate3D(position[0], position[1], position[2]).Mul4(mr.LocalTransform.Matrix())

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
		_ = mr.Shader.SetUniformVec4(shaders.UniformSrcRect, mgl32.Vec4{0.0, 1.0, 1.0, 1.0})
	}

	if len(mr.Group) == 0 {
		mr.Mesh.DrawAll()
	} else {
		mr.Mesh.DrawGroup(mr.Group)
	}
}
