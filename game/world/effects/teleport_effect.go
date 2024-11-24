package effects

import (
	"math/rand/v2"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

func Teleport(radius float32) comps.ParticleRender {
	poofTexture := cache.GetTexture("assets/textures/sprites/teleport_poof.png")
	return comps.ParticleRender{
		Mesh:             cache.QuadMesh,
		Texture:          poofTexture,
		SpawnRate:        0.01,
		SpawnRadius:      radius,
		BurstCount:       20,
		VisibilityRadius: 5.0,
		EmissionTimer:    0.0,
		MaxCount:         20,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = mgl32.Vec4{0.25 + rand.Float32()*0.75, 0.25, 0.5 + rand.Float32()*0.5, 1.0}
			s := rand.Float32()*0.20 + 0.20
			form.Size = mgl32.Vec2{s, s}
			info.Velocity = info.Velocity.Mul(rand.Float32()*2 + 1.0)
			info.Acceleration = mgl32.Vec3{}
			info.Lifetime = 0.5
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
