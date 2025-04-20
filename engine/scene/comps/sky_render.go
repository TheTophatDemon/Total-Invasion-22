package comps

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type SkyRender struct {
	Mesh    *geom.Mesh
	Shader  *shaders.Shader
	Texture *textures.Texture
}

func NewSkyRender(Mesh *geom.Mesh, Shader *shaders.Shader, Texture *textures.Texture) SkyRender {
	return SkyRender{
		Mesh,
		Shader,
		Texture,
	}
}

// Renders the mesh with the given local transform and the optional animation player.
// If transform is nil, then the model matrix is the identity matrix.
func (sr *SkyRender) Render(
	context *render.Context,
) {
	if sr.Mesh == nil || sr.Shader == nil {
		return
	}

	// Bind resources
	sr.Mesh.Bind()
	shaders.SkyShader.Use()
	if sr.Texture != nil {
		sr.Texture.Bind()
	}

	// Set uniforms
	_ = context.SetUniforms(sr.Shader)
	_ = sr.Shader.SetUniformInt(shaders.UniformTex, 0)

	gl.DepthMask(false)
	defer gl.DepthMask(true)

	sr.Mesh.DrawAll()
}
