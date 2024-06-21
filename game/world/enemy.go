package world

import (
	"math"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const (
	ENEMY_FOV_RADS       = math.Pi
	ENEMY_WAKE_PROXIMITY = 1.5
)

type Enemy struct {
	SpriteRender         comps.SpriteRender
	AnimPlayer           comps.AnimationPlayer
	bloodParticles       comps.ParticleRender
	actor                Actor
	world                *World
	walkAnim, attackAnim textures.Animation
}

var _ HasActor = (*Enemy)(nil)
var _ comps.HasBody = (*Enemy)(nil)

func (enemy *Enemy) Actor() *Actor {
	return &enemy.actor
}

func (enemy *Enemy) Body() *comps.Body {
	return &enemy.actor.body
}

func SpawnEnemy(storage *scene.Storage[Enemy], position, angles mgl32.Vec3, world *World) (id scene.Id[*Enemy], enemy *Enemy, err error) {
	id, enemy, err = storage.New()
	if err != nil {
		return
	}

	enemy.world = world

	wraithTexture := cache.GetTexture("assets/textures/sprites/wraith.png")
	enemy.walkAnim, _ = wraithTexture.GetAnimation("walk;front")
	enemy.attackAnim, _ = wraithTexture.GetAnimation("walk;front")

	enemy.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(
				mgl32.Vec3(position).Add(mgl32.Vec3{0.0, -0.1, 0.0}), math2.DegToRadVec3(angles), mgl32.Vec3{0.9, 0.9, 0.9},
			),
			Shape:  collision.NewSphere(0.7),
			Layer:  COL_LAYER_ACTORS,
			Filter: COL_FILTER_FOR_ACTORS,
			LockY:  true,
		},
		YawAngle:  mgl32.DegToRad(angles[1]),
		AccelRate: 80.0,
		Friction:  20.0,
		MaxSpeed:  5.0,
	}
	enemy.SpriteRender = comps.NewSpriteRender(wraithTexture)
	enemy.AnimPlayer = comps.NewAnimationPlayer(enemy.walkAnim, false)

	bloodTexture := cache.GetTexture("assets/textures/sprites/blood.png")
	bloodAnim, _ := bloodTexture.GetAnimation("default")
	enemy.bloodParticles = comps.ParticleRender{
		Mesh:             cache.QuadMesh,
		Texture:          bloodTexture,
		SpawnRate:        0.01,
		SpawnRadius:      0.5,
		VisibilityRadius: 5.0,
		EmissionTimer:    0.0,
		SpawnFunc: func(index int, form *comps.ParticleForm, info *comps.ParticleInfo) {
			form.Color = color.Red.Vector()
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
	enemy.bloodParticles.Init(25)

	return
}

func (enemy *Enemy) Update(deltaTime float32) {
	enemy.AnimPlayer.Update(deltaTime)
	enemy.actor.Update(deltaTime)
	enemy.bloodParticles.Update(deltaTime, &enemy.Body().Transform)

	enemyPos := enemy.Body().Transform.Position()
	enemyDir := enemy.actor.FacingVec()

	if player, ok := enemy.world.CurrentPlayer.Get(); ok {
		toPlayer := player.Body().Transform.Position().Sub(enemyPos)
		distToPlayer := toPlayer.Len()
		if distToPlayer != 0.0 {
			toPlayer = toPlayer.Normalize()
		}
		angle := math2.Acos(toPlayer.Dot(enemyDir))

		enemy.AnimPlayer.Stop()
		if angle < ENEMY_FOV_RADS/2.0 || distToPlayer < ENEMY_WAKE_PROXIMITY {
			res, handle := enemy.world.Raycast(enemyPos, toPlayer, COL_LAYER_ACTORS, 100.0, enemy)
			if handle.Equals(player.id.Handle) && res.Hit {
				enemy.AnimPlayer.Play()
			}
		}
	}
}

func (enemy *Enemy) Render(context *render.Context) {
	enemy.SpriteRender.Render(&enemy.Body().Transform, &enemy.AnimPlayer, context, enemy.actor.YawAngle)
	enemy.bloodParticles.Render(&enemy.Body().Transform, context)
}

func (enemy *Enemy) ProcessSignal(signal Signal, params any) {
}
