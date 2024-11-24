package effects

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

func Explosion(maxCount int, rate float32, radius float32) comps.ParticleRender {
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
