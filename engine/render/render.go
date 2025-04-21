package render

import (
	"errors"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

// Contains global information useful for rendering.
type Context struct {
	View, ViewInverse                                     mgl32.Mat4
	Projection                                            mgl32.Mat4
	FogStart                                              float32
	FogLength                                             float32
	LightDirection                                        mgl32.Vec3
	AmbientColor                                          mgl32.Vec3
	AspectRatio                                           float32
	DrawnSpriteCount, DrawnWallCount, DrawnParticlesCount uint32

	viewProjection mgl32.Mat4
	cameraFrustum  math2.Frustum
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

func (context *Context) ViewProjection() mgl32.Mat4 {
	if context.viewProjection.Trace() == 0.0 {
		context.viewProjection = context.Projection.Mul4(context.View)
	}
	return context.viewProjection
}

func (context *Context) CameraFrustum() math2.Frustum {
	if context.cameraFrustum.Planes[0].Normal.LenSqr() == 0.0 {
		context.cameraFrustum = math2.FrustumFromMatrices(context.ViewProjection().Inv())
	}
	return context.cameraFrustum
}

func IsPointVisible(context *Context, point mgl32.Vec3) bool {
	return context.CameraFrustum().ContainsPoint(point)
}

func IsBoxVisible(context *Context, box math2.Box) bool {
	return context.CameraFrustum().IntersectsBox(box)
}

func IsSphereVisible(context *Context, point mgl32.Vec3, radius float32) bool {
	return context.CameraFrustum().IntersectsSphere(point, radius)
}
