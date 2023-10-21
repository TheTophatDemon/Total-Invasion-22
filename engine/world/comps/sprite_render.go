package comps

import (
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
	cameraPos := context.View.Inv().Col(3).Vec3()
	toCamera := cameraPos.Sub(transform.Position())
	if toCamera.LenSqr() > mgl32.Epsilon {
		toCamera = toCamera.Normalize()
		yawMatrix := mgl32.Rotate3DY(yawAngle)
		radAngleDiff := math2.Acos(toCamera.Dot(mgl32.TransformCoordinate(math2.Vec3Forward(), yawMatrix.Mat4())))
		angleDifference := mgl32.RadToDeg(radAngleDiff)
		layer, _, found := sr.meshRender.Texture.FindLayerWithinAngle(int(angleDifference))
		if found {
			anim, found := sr.meshRender.Texture.GetAnimation(animPlayer.animation.BaseName() + ";" + layer.Name)
			if found {
				animPlayer.SwapAnimation(anim)
			}
		}
	}
	sr.meshRender.Render(transform, animPlayer, context)
}
