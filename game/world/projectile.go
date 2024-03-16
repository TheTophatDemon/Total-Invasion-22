package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type Projectile struct {
	id           scene.Id[*Projectile]
	SpriteRender comps.SpriteRender
	AnimPlayer   comps.AnimationPlayer
	body         comps.Body
	owner        scene.Handle
	moveFunc     func(deltaTime float32)
	speed        float32
}

var _ comps.HasBody = (*Projectile)(nil)

func (proj *Projectile) Body() *comps.Body {
	return &proj.body
}

func SpawnSickle(st *scene.Storage[Projectile], position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = st.New()

	proj.id = id
	proj.owner = owner

	proj.body.Transform = comps.TransformFromTranslationAngles(position, rotation)
	proj.body.Shape = collision.NewSphere(0.5)
	proj.body.Layer = COL_LAYER_PROJECTILES
	proj.body.Filter = COL_LAYER_NONE
	proj.body.OnIntersect = proj.OnIntersect

	sickleTex := cache.GetTexture("assets/textures/sprites/sickle_thrown.png")
	proj.SpriteRender = comps.NewSpriteRender(sickleTex)

	throwAnim, ok := sickleTex.GetAnimation("throw;front")
	if !ok {
		log.Println("could not find animation for thrown sickle sprite")
	}
	proj.AnimPlayer = comps.NewAnimationPlayer(throwAnim, true)
	proj.speed = 35.0

	proj.moveFunc = func(deltaTime float32) {
		proj.speed = max(-35.0, proj.speed-deltaTime*50.0)
		if owner, ok := scene.Get[HasActor](proj.owner); ok && proj.speed < 0.0 {
			//TODO: Set rotation to face player
			ownerPos := owner.Body().Transform.Position()
			projPos := proj.body.Transform.Position()
			proj.body.Transform.SetRotation(0.0, math2.Atan2(projPos.Z()-ownerPos.Z(), ownerPos.X()-projPos.X())+math2.HALF_PI, 0.0)
		}
		proj.body.Velocity = mgl32.TransformNormal(mgl32.Vec3{0.0, 0.0, -proj.speed}, proj.body.Transform.Matrix())
	}

	return
}

func (proj *Projectile) Update(deltaTime float32) {
	proj.AnimPlayer.Update(deltaTime)
	if proj.moveFunc != nil {
		proj.moveFunc(deltaTime)
	}
}

func (proj *Projectile) Render(context *render.Context) {
	proj.SpriteRender.Render(&proj.body.Transform, &proj.AnimPlayer, context, proj.body.Transform.Yaw())
}

func (proj *Projectile) OnIntersect(body *comps.Body, result collision.Result) {
	if proj.speed < 0.0 {
		if owner, ok := scene.Get[HasActor](proj.owner); ok && body == owner.Body() {
			proj.id.Remove()
		}
	}
	if body.Layer&COL_LAYER_MAP != 0 && proj.speed > 0.0 {
		proj.speed = -proj.speed / 2.0
	}
}
