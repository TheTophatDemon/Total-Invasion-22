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

type SpriteRender struct {
	meshRender       MeshRender
	DiffuseColor     color.Color
	AdditiveBlending bool
}

func NewSpriteRender(texture *textures.Texture) SpriteRender {
	return SpriteRender{
		meshRender: NewMeshRender(
			cache.QuadMesh,
			shaders.SpriteShader,
			texture,
		),
		DiffuseColor: color.White,
	}
}

func NewSpriteRenderWithColor(texture *textures.Texture, diffuseColor color.Color) SpriteRender {
	return SpriteRender{
		meshRender: NewMeshRender(
			cache.QuadMesh,
			shaders.SpriteShader,
			texture,
		),
		DiffuseColor: diffuseColor,
	}
}

func (sr *SpriteRender) Texture() *textures.Texture {
	return sr.meshRender.Texture
}

func (sr *SpriteRender) Render(
	transform *Transform,
	animPlayer *AnimationPlayer,
	context *render.Context,
	yawAngle float32,
) bool {
	if !context.IsSphereVisible(transform.Position(), transform.Scale().X()) {
		return false
	}

	if sr.meshRender.Shader == nil {
		return false
	}
	sr.meshRender.Shader.Use()

	if sr.meshRender.Texture != nil && sr.meshRender.Texture.LayerCount() > 1 {
		// Change animation layer based on angle to the camera
		cameraPos := context.ViewInverse.Col(3).Vec3()
		toCamera := cameraPos.Sub(transform.Position())
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
			layer, flip, found := sr.meshRender.Texture.FindLayerToDisplay(angleDifference, settings.Current.Locale)
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

	if sr.AdditiveBlending {
		gl.BlendFunc(gl.ONE, gl.ONE)
		defer gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	}
	sr.meshRender.Render(transform, animPlayer, context)

	context.DrawnSpriteCount++
	return true
}
