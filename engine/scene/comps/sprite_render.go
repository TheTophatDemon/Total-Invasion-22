package comps

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	SpriteRender struct {
		meshRender   MeshRender
		DiffuseColor color.Color
	}
)

func NewSpriteRender(
	texture *textures.Texture, diffuseColor *color.Color, size *mgl32.Vec2,
) SpriteRender {
	sr := SpriteRender{
		meshRender: NewMeshRender(
			cache.QuadMesh,
			shaders.SpriteShader,
			texture,
		),
	}

	if size != nil {
		sr.SetScale(size[0], size[1])
	}

	if diffuseColor != nil {
		sr.DiffuseColor = *diffuseColor
	} else {
		sr.DiffuseColor = color.White
	}

	return sr
}

func (sr *SpriteRender) Texture() *textures.Texture {
	return sr.meshRender.Texture
}

func (sr *SpriteRender) SetScale(scaleX, scaleY float32) {
	if sr == nil {
		return
	}
	sr.meshRender.LocalTransform.SetScale(scaleX, scaleY, 1.0)
}

func (sr SpriteRender) Scale() mgl32.Vec2 {
	return sr.meshRender.LocalTransform.Scale().Vec2()
}

func (sr *SpriteRender) Render(
	position mgl32.Vec3,
	animPlayer *AnimationPlayer,
	context *render.Context,
	yawAngle float32,
) bool {
	if sr.meshRender.Shader == nil {
		return false
	}

	if !context.DrawingTransparent && !context.IsSphereVisible(position, max(sr.Scale()[0], sr.Scale()[1])) {
		return false
	}

	blendAdd := sr.Texture() != nil && sr.Texture().HasFlag(textures.FlagBlendAdd)

	if (sr.DiffuseColor.A < 1.0 || blendAdd) && !context.DrawingTransparent {
		//TODO: Optimize?
		// Avoid making extra matrix
		// Abstract render parameters into generic struct so that you can avoid closure captures
		context.EnqueueTransparentRender(func(context *render.Context) {
			sr.Render(position, animPlayer, context, yawAngle)
		}, context.DistanceFromScreen(mgl32.Translate3D(position[0], position[1], position[2])))
		return true
	}

	sr.meshRender.Shader.Use()

	if sr.meshRender.Texture != nil && sr.meshRender.Texture.LayerCount() > 1 {
		// Change animation layer based on angle to the camera
		cameraPos := context.ViewInverse.Col(3).Vec3()
		toCamera := cameraPos.Sub(position)
		if toCamera.LenSqr() > mgl32.Epsilon {
			toCamera = toCamera.Normalize()
			ourDirection := mgl32.TransformCoordinate(math2.Vec3Forward(), mgl32.Rotate3DY(yawAngle).Mat4())
			dp := toCamera.Dot(ourDirection)
			cross := toCamera.Cross(ourDirection)
			radAngleDiff := math2.Acos(dp)
			angleDifference := int(mgl32.RadToDeg(radAngleDiff))
			if cross.Dot(math2.Vec3Up()) < 0.0 {
				angleDifference *= -1
			}
			layer, flip, found := sr.meshRender.Texture.FindLayerToDisplay(angleDifference, string(settings.Current.Locale))
			if found {
				anim, found := sr.meshRender.Texture.GetAnimation(animPlayer.animation.BaseName() + ";" + layer.Name)
				if found {
					animPlayer.SwapAnimation(anim)
				}
				err := sr.meshRender.Shader.SetUniformBool(shaders.UniformFlipHorz, flip)
				if err != nil {
					failure.LogErrWithLocation("error setting uniform: %v", err)
				}
			}
		}
	} else {
		err := sr.meshRender.Shader.SetUniformBool(shaders.UniformFlipHorz, false)
		if err != nil {
			failure.LogErrWithLocation("error setting uniform: %v", err)
		}
	}

	err := sr.meshRender.Shader.SetUniformVec4(shaders.UniformDiffuseColor, sr.DiffuseColor.Vector())
	if err != nil {
		failure.LogErrWithLocation("error setting diffuse color uniform: %v", err)
	}

	if blendAdd {
		gl.BlendFunc(gl.ONE, gl.ONE)
		defer gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	}
	sr.meshRender.Render(position, animPlayer, context)

	context.DrawnSpriteCount++
	return true
}
