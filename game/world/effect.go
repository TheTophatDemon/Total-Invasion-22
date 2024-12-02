package world

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game/settings"
	"tophatdemon.com/total-invasion-ii/game/world/effects"
)

type Effect struct {
	Transform comps.Transform
	Lifetime  float32

	world     *World
	id        scene.Id[*Effect]
	particles comps.ParticleRender
	voice     tdaudio.VoiceId
}

func SpawnEffect(world *World, transform comps.Transform, lifetime float32, particles comps.ParticleRender) (id scene.Id[*Effect], fx *Effect, err error) {
	id, fx, err = world.Effects.New()
	if err != nil {
		return
	}

	*fx = Effect{
		Transform: transform,
		Lifetime:  lifetime,
		world:     world,
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
	fx.Lifetime -= deltaTime
	fx.particles.Update(deltaTime, &fx.Transform)
	if fx.voice.IsValid() {
		fx.voice.SetPositionV(fx.Transform.Position())
	}
	if fx.Lifetime < 0.0 {
		fx.world.QueueRemoval(fx.id.Handle)
	}
}

func (fx *Effect) Render(context *render.Context) {
	fx.particles.Render(&fx.Transform, context)
}

func SpawnSingleExplosion(world *World, transform comps.Transform) (id scene.Id[*Effect], fx *Effect, err error) {
	const DAMAGE_RADIUS = 3.5
	const MAX_ENEMY_DAMAGE = 175.0
	const MIN_ENEMY_DAMAGE = 50.0
	id, fx, err = SpawnEffect(world, transform, 1.0, effects.Explosion(1, 1.0, 1.5))
	if err != nil {
		return
	}
	fx.voice = cache.GetSfx("assets/sounds/explosion.wav").PlayAttenuatedV(transform.Position())

	// Apply splash damage to surrounding entities
	for _, handle := range world.BodiesInSphere(transform.Position(), DAMAGE_RADIUS, nil) {
		if bodyHaver, ok := scene.Get[comps.HasBody](handle); ok {
			if damageable, ok := bodyHaver.(Damageable); ok {
				vecToTarget := bodyHaver.Body().Transform.Position().Sub(transform.Position())
				distanceToExplosion := vecToTarget.Len()
				if distanceToExplosion > 0 {
					cast, _ := world.Raycast(transform.Position(), vecToTarget.Mul(1.0/distanceToExplosion),
						COL_LAYER_MAP, distanceToExplosion, nil)
					// Do not apply damage to entities when there is a wall between them and the explosion.
					if cast.Hit {
						continue
					}
				}
				if sphere, isSphere := bodyHaver.Body().Shape.(collision.Sphere); isSphere {
					distanceToExplosion -= sphere.Radius()
				}
				var damage float32
				if _, isPlayer := damageable.(*Player); isPlayer {
					difficulty := settings.CurrDifficulty()
					damage = math2.Lerp(difficulty.ExplosionMaxDamage, difficulty.ExplosionMinDamage, distanceToExplosion/DAMAGE_RADIUS)
				} else {
					damage = math2.Lerp(MAX_ENEMY_DAMAGE, MIN_ENEMY_DAMAGE, distanceToExplosion/DAMAGE_RADIUS)
				}
				damageable.OnDamage(fx, damage)
			}
		}
	}
	return
}
