package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/audio"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

func SpawnEgg(world *World, st *scene.Storage[Projectile], position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = st.New()
	if err != nil {
		return
	}

	proj.world = world
	proj.id = id
	proj.owner = owner

	proj.body = comps.Body{
		Transform:      comps.TransformFromTranslationAnglesScale(position, rotation, mgl32.Vec3{0.4, 0.4, 0.4}),
		Shape:          collision.NewSphere(0.1),
		Layer:          COL_LAYER_PROJECTILES,
		Filter:         COL_LAYER_MAP | COL_LAYER_ACTORS,
		LockY:          true,
		SweepCollision: true,
	}

	eggTex := cache.GetTexture("assets/textures/sprites/egg.png")
	proj.SpriteRender = comps.NewSpriteRender(eggTex)
	proj.speed = 100.0

	var sfxShoot *audio.Sfx
	sfxShoot, err = cache.GetSfx("assets/sounds/chickengun.wav")
	if err != nil {
		log.Println(err)
	} else {
		proj.voices[0] = sfxShoot.Play()
	}

	proj.moveFunc = proj.eggMove
	proj.body.OnIntersect = proj.eggIntersect

	return
}

func (proj *Projectile) eggMove(deltaTime float32) {
	proj.body.Velocity = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -proj.speed}, proj.body.Transform.Matrix())
}

func (proj *Projectile) eggIntersect(body *comps.Body, result collision.Result) {
	if owner, ok := scene.Get[HasActor](proj.owner); ok && body != owner.Body() && body.OnLayer(proj.body.Filter) {
		proj.id.Remove()
	}
}
