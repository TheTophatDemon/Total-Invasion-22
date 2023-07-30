package render

import (
	"errors"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

// Contains global information useful for rendering.
type Context struct {
	View           mgl32.Mat4
	Projection     mgl32.Mat4
	FogStart       float32
	FogLength      float32
	LightDirection mgl32.Vec3
	AmbientColor   mgl32.Vec3
}

func (context *Context) SetUniforms(shader *assets.Shader) error {
	return errors.Join(shader.SetUniformMatrix(assets.UniformViewMatrix, context.View),
		shader.SetUniformMatrix(assets.UniformProjMatrix, context.Projection),
		shader.SetUniformFloat(assets.UniformFogStart, context.FogStart),
		shader.SetUniformFloat(assets.UniformFogLength, context.FogLength),
		shader.SetUniformVec3(assets.UniformLightDir, context.LightDirection),
		shader.SetUniformVec3(assets.UniformAmbientColor, context.AmbientColor))
}
