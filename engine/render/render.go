package render

import (
	"errors"
	"slices"

	"github.com/go-gl/gl/v3.3-core/gl"
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
	DrawingTransparent                                    bool

	viewProjection mgl32.Mat4
	cameraFrustum  math2.Frustum
	transpQueue    []transpRender // Queue for rendering translucent objects after opaque objects.
}

type transpRender struct {
	renderFunc func(context *Context)
	distance   float32 // Distance from the camera / value of Z axis in screen space.
}

func (context *Context) Enable3D() {
	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.DepthFunc(gl.LESS)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.SCISSOR_TEST)
	gl.CullFace(gl.BACK)
}

func (context *Context) Enable2D() {
	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.FRONT)
	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.DepthFunc(gl.LESS)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Disable(gl.SCISSOR_TEST)
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

func (context *Context) IsBoxVisible(box math2.Box) bool {
	return context.CameraFrustum().IntersectsBox(box)
}

func (context *Context) IsSphereVisible(point mgl32.Vec3, radius float32) bool {
	return context.CameraFrustum().IntersectsSphere(point, radius)
}

func (context *Context) DistanceFromScreen(modelMatrix mgl32.Mat4) float32 {
	return mgl32.TransformCoordinate(
		mgl32.Vec3{0.0, 0.0, 1.0},
		context.ViewInverse.Mul4(modelMatrix),
	)[2]
}

func (context *Context) EnqueueTransparentRender(renderFunc func(context *Context), distance float32) {
	if renderFunc == nil {
		return
	}
	i := 0
	// Need a boomer loop here so that i can get set to len(context.transpQueue)
	for ; i < len(context.transpQueue); i += 1 {
		// Maintains sorted order from farthest to nearest towards the camera plane.
		if context.transpQueue[i].distance < distance {
			break
		}
	}
	context.transpQueue = slices.Insert(context.transpQueue, i, transpRender{renderFunc, distance})
}

func (context *Context) RenderTranslucentObjects() {
	gl.DepthMask(false)
	defer gl.DepthMask(true)
	context.DrawingTransparent = true
	defer func() { context.DrawingTransparent = false }()
	for _, tRender := range context.transpQueue {
		tRender.renderFunc(context)
	}
	context.transpQueue = context.transpQueue[0:0]
}
