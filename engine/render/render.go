package render

import (
	"errors"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
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

func (context *Context) SetUniforms(shader *shaders.Shader) error {
	return errors.Join(shader.SetUniformMatrix(shaders.UniformViewMatrix, context.View),
		shader.SetUniformMatrix(shaders.UniformProjMatrix, context.Projection),
		shader.SetUniformFloat(shaders.UniformFogStart, context.FogStart),
		shader.SetUniformFloat(shaders.UniformFogLength, context.FogLength),
		shader.SetUniformVec3(shaders.UniformLightDir, context.LightDirection),
		shader.SetUniformVec3(shaders.UniformAmbientColor, context.AmbientColor))
}
