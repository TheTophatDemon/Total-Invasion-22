package render

import (
	"errors"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
)

// Contains global information useful for rendering.
type Context struct {
	View             mgl32.Mat4
	Projection       mgl32.Mat4
	FogStart         float32
	FogLength        float32
	LightDirection   mgl32.Vec3
	AmbientColor     mgl32.Vec3
	DrawnSpriteCount uint32
}

func (context *Context) SetUniforms(shader *shaders.Shader) error {
	return errors.Join(
		shader.SetUniformMatrix(shaders.UniformViewMatrix, context.View),
		shader.SetUniformMatrix(shaders.UniformProjMatrix, context.Projection),
		shader.SetUniformFloat(shaders.UniformFogStart, context.FogStart),
		shader.SetUniformFloat(shaders.UniformFogLength, context.FogLength),
		shader.SetUniformVec3(shaders.UniformLightDir, context.LightDirection),
		shader.SetUniformVec3(shaders.UniformAmbientColor, context.AmbientColor),
	)
}

func IsPointVisible(context *Context, point mgl32.Vec3) bool {
	var screenSpacePoint mgl32.Vec3 = mgl32.TransformCoordinate(point, context.Projection.Mul4(context.View))
	return (screenSpacePoint.Z() > -1.0 && screenSpacePoint.Z() < 1.0 &&
		screenSpacePoint.X() > -1.0 && screenSpacePoint.X() < 1.0 &&
		screenSpacePoint.Y() > -1.0 && screenSpacePoint.Y() < 1.0)
}

func IsSphereVisible(context *Context, point mgl32.Vec3, radius float32) bool {
	var projPoint mgl32.Vec4 = context.Projection.Mul4(context.View).Mul4x1(point.Vec4(1.0))
	radius /= projPoint.W() * 0.5
	screenSpacePoint := projPoint.Vec3().Mul(1.0 / projPoint.W())
	return (screenSpacePoint.Z()+radius > -1.0 && screenSpacePoint.Z()-radius < 1.0 &&
		screenSpacePoint.X()+radius > -1.0 && screenSpacePoint.X()-radius < 1.0 &&
		screenSpacePoint.Y()+radius > -1.0 && screenSpacePoint.Y()-radius < 1.0)
}
