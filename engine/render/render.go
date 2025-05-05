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
	DrawingTranslucent                                    bool

	viewProjection   mgl32.Mat4
	cameraFrustum    math2.Frustum
	translucentQueue []translucentPair // Queue for rendering translucent objects after opaque objects.
}

type translucentPair struct {
	TranslucentRender
	distance float32 // Distance from the camera / value of Z axis in screen space.
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

func (context *Context) EnqueueTranslucentRender(item TranslucentRender) {
	distance := item.DistanceFromScreen(context)
	for i := range context.translucentQueue {
		// Maintains sorted order from farthest to nearest towards the camera plane.
		if context.translucentQueue[i].distance < distance {
			context.translucentQueue = slices.Insert(context.translucentQueue, i, translucentPair{
				TranslucentRender: item,
				distance:          distance,
			})
			return
		}
	}
	context.translucentQueue = append(context.translucentQueue, translucentPair{
		TranslucentRender: item,
		distance:          distance,
	})
}

func (context *Context) RenderTranslucentObjects() {
	gl.DepthMask(false)
	defer gl.DepthMask(true)
	context.DrawingTranslucent = true
	defer func() { context.DrawingTranslucent = false }()
	for _, obj := range context.translucentQueue {
		obj.Render(context)
	}
	context.translucentQueue = context.translucentQueue[0:0]
}
