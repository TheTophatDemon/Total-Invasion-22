package comps

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type SpriteRender struct {
	meshRender MeshRender
}

func NewSpriteRender(texture *textures.Texture) SpriteRender {
	return SpriteRender{
		meshRender: NewMeshRender(
			cache.QuadMesh,
			shaders.SpriteShader,
			texture,
		),
	}
}

func (sr *SpriteRender) Render(
	transform *Transform,
	animPlayer *AnimationPlayer,
	context *render.Context,
	yawAngle float32,
) {
	if !render.IsSphereVisible(context, transform.Position(), transform.Scale().X()) {
		return
	}

	sr.meshRender.Shader.Use()

	if sr.meshRender.Texture != nil && sr.meshRender.Texture.LayerCount() > 1 {
		// Change animation layer based on angle to the camera
		cameraPos := context.View.Inv().Col(3).Vec3()
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
			layer, flip, found := sr.meshRender.Texture.FindLayerWithinAngle(angleDifference)
			if found {
				anim, found := sr.meshRender.Texture.GetAnimation(animPlayer.animation.BaseName() + ";" + layer.Name)
				if found {
					animPlayer.SwapAnimation(anim)
				}
				err := sr.meshRender.Shader.SetUniformBool(shaders.UniformFlipHorz, flip)
				if err != nil {
					log.Println("Error setting uniform in (*SpriteRender).Render", err)
				}
			}
		}
	} else {
		err := sr.meshRender.Shader.SetUniformBool(shaders.UniformFlipHorz, false)
		if err != nil {
			log.Println("Error setting uniform in (*SpriteRender).Render", err)
		}
	}

	sr.meshRender.Render(transform, animPlayer, context)

	context.DrawnSpriteCount++
}
