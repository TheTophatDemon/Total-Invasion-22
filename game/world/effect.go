package world

import (
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Effect struct {
	Position mgl32.Vec3
	Lifetime float32

	id        scene.Id[*Effect]
	particles comps.ParticleRender
	voice     tdaudio.VoiceId
}

func SpawnEffect(position mgl32.Vec3, lifetime float32, particles comps.ParticleRender) (id scene.Id[*Effect], fx *Effect, err error) {
	id, fx, err = gWorld.Effects.New()
	if err != nil {
		return
	}

	*fx = Effect{
		Position:  position,
		Lifetime:  lifetime,
		id:        id,
		particles: particles,
	}

	fx.particles.Init()
	return
}

func (fx *Effect) Finalize() {
	fx.particles.Finalize()
}

func (fx *Effect) Update(deltaTime float32) {
	if fx == nil {
		return
	}
	fx.Lifetime -= deltaTime
	fx.particles.Update(deltaTime, fx.Position)
	if fx.voice.IsValid() {
		fx.voice.SetPositionV(fx.Position)
	}
	if fx.Lifetime < 0.0 {
		gWorld.QueueRemoval(fx.id.Handle)
	}
}

func (fx *Effect) Render(context *render.Context) {
	if fx == nil {
		return
	}
	fx.particles.Render(fx.Position, context)
}

func SpawnSingleExplosion(position mgl32.Vec3) (id scene.Id[*Effect], fx *Effect, err error) {
	const damageRadius = 3.5
	const maxEnemyDamage = 200.0
	const minEnemyDamage = 50.0
	id, fx, err = SpawnEffect(position, 1.0, ExplosionParticles(1, 1.0, 1.5, 0.0))
	if err != nil {
		return
	}
	fx.voice = cache.GetSfx("assets/sounds/explosion.wav").PlayAttenuatedV(position)

	// Apply splash damage to surrounding entities
	actorIter := gWorld.IterActorsInSphere(position, damageRadius, nil)
	for actorHaver, _ := actorIter.Next(); actorHaver != nil; actorHaver, _ = actorIter.Next() {
		if damageable, ok := actorHaver.(Damageable); ok {
			vecToTarget := actorHaver.Body().Position.Sub(position)
			distanceToExplosion := vecToTarget.Len()
			if distanceToExplosion > 0 {
				cast, _ := gWorld.Raycast(position, vecToTarget.Mul(1.0/distanceToExplosion),
					ColLayerMap, distanceToExplosion, nil)
				// Do not apply damage to entities when there is a wall between them and the explosion.
				if cast.Hit {
					continue
				}
			}
			distanceToExplosion -= actorHaver.Body().Shape.Extents().Max[0]
			var damage float32
			if _, isPlayer := damageable.(*Player); isPlayer {
				difficulty := settings.CurrDifficulty()
				damage = math2.Lerp(difficulty.ExplosionMaxDamage, difficulty.ExplosionMinDamage, distanceToExplosion/damageRadius)
			} else {
				damage = math2.Lerp(minEnemyDamage, maxEnemyDamage, 1.0-(distanceToExplosion/damageRadius))
			}
			damageable.OnDamage(fx, damage)
		}
	}
	return
}

func BloodParticles(maxCount int, color color.Color, radius float32) comps.ParticleRender {
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
			const shrinkRate = 0.75
			form.Size[0] -= deltaTime * shrinkRate
			form.Size[1] -= deltaTime * shrinkRate
			if form.Size[0] <= 0.1 {
				form.Size = mgl32.Vec2{}
				info.Lifetime = 0.0
			}
		},
	}
}

func EggParticles(radius float32) comps.ParticleRender {
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
			const shrinkRate = 0.75
			form.Size[0] -= deltaTime * shrinkRate
			form.Size[1] -= deltaTime * shrinkRate
			if form.Size[0] <= 0.1 {
				form.Size = mgl32.Vec2{}
				info.Lifetime = 0.0
			}
		},
	}
}

func ExplosionParticles(maxCount int, rate float32, spriteSize, spawnRadius float32) comps.ParticleRender {
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
		SpawnRadius:   spawnRadius,
		SpawnRate:     rate,
		BurstCount:    1,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = color.White.Vector()
			form.Size = mgl32.Vec2{spriteSize, spriteSize}
			info.Velocity = mgl32.Vec3{}
			info.Acceleration = mgl32.Vec3{}
			info.Lifetime = 1.0
			info.AnimPlayer = comps.NewAnimationPlayer(anim, true)
		},
	}
}

func TeleportParticles(radius float32) comps.ParticleRender {
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
			const shrinkRate = 0.75
			form.Size[0] -= deltaTime * shrinkRate
			form.Size[1] -= deltaTime * shrinkRate
			if form.Size[0] <= 0.1 {
				form.Size = mgl32.Vec2{}
				info.Lifetime = 0.0
			}
		},
	}
}
