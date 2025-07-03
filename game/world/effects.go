package world

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

func BloodEffect(maxCount int, color color.Color, radius float32) comps.ParticleRender {
	bloodTexture := cache.GetTexture("assets/textures/sprites/blood.png")
	bloodAnim, _ := bloodTexture.GetAnimation("default")
	return comps.ParticleRender{
		Mesh:             cache.QuadMesh,
		Texture:          bloodTexture,
		SpawnRate:        0.01,
		SpawnRadius:      radius,
		VisibilityRadius: 5.0,
		EmissionTimer:    0.0,
		MaxCount:         maxCount,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = color.Vector()
			s := rand.Float32()*0.10 + 0.15
			form.Size = mgl32.Vec2{s, s}
			info.Velocity = info.Velocity.Mul(rand.Float32()*5 + 1.0)
			info.Acceleration = mgl32.Vec3{0.0, -20.0, 0.0}
			info.Lifetime = 1.0
			info.AnimPlayer = comps.NewAnimationPlayer(bloodAnim, false)
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

func EggShardsEffect(radius float32) comps.ParticleRender {
	particleTex := cache.GetTexture("assets/textures/sprites/egg_shards.png")
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

func ExplosionEffect(maxCount int, rate float32, radius float32) comps.ParticleRender {
	tex := cache.GetTexture("assets/textures/sprites/explosion.png")
	anim := tex.GetDefaultAnimation()
	if rate == 0 {
		rate = 1.0
	}
	return comps.ParticleRender{
		Mesh:          cache.QuadMesh,
		Texture:       tex,
		EmissionTimer: 0.2,
		MaxCount:      maxCount,
		SpawnRadius:   radius,
		SpawnRate:     rate,
		BurstCount:    1,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = color.White.Vector()
			form.Size = mgl32.Vec2{radius, radius}
			info.Velocity = mgl32.Vec3{}
			info.Acceleration = mgl32.Vec3{}
			info.Lifetime = 1.0
			info.AnimPlayer = comps.NewAnimationPlayer(anim, true)
		},
	}
}

func TeleportEffect(radius float32) comps.ParticleRender {
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
