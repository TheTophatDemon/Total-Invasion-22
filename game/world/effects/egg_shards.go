package effects

import (
	"math/rand/v2"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const TEX_EGG_SHARDS = "assets/textures/sprites/egg_shards.png"

func EggShards(radius float32) comps.ParticleRender {
	particleTex := cache.GetTexture(TEX_EGG_SHARDS)
	return comps.ParticleRender{
		Texture:       particleTex,
		EmissionTimer: 0.2,
		MaxCount:      4,
		SpawnRadius:   radius,
		SpawnRate:     1.0,
		BurstCount:    6,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = color.White.Vector()
			form.Size = mgl32.Vec2{0.2, 0.2}
			info.Velocity = info.Velocity.Mul(rand.Float32()*1 + 5.0)
			info.Acceleration = mgl32.Vec3{0.0, -20.0, 0.0}
			info.Lifetime = 1.0
			info.AnimPlayer = comps.NewAnimationPlayer(particleTex.GetDefaultAnimation(), false)
			info.AnimPlayer.MoveToRandomFrame()
		},
		UpdateFunc: func(deltaTime float32, form *comps.ParticleForm, info *comps.ParticleInfo) {
			const SHRINK_RATE = 0.75
			form.Size[0] -= deltaTime * SHRINK_RATE
			form.Size[1] -= deltaTime * SHRINK_RATE
			if form.Size[0] <= 0.1 {
				form.Size = mgl32.Vec2{}
				info.Lifetime = 0.0
			}
		},
	}
}
