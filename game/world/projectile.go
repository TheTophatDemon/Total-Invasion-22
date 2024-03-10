package world

import (
	"fmt"
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
	SpriteRender comps.SpriteRender
	AnimPlayer   comps.AnimationPlayer
	body         comps.Body
	owner        scene.Handle
	moveFunc     func(deltaTime float32)
}

var _ comps.HasBody = (*Projectile)(nil)

func (proj *Projectile) Body() *comps.Body {
	return &proj.body
}

func SpawnSickle(st *scene.Storage[Projectile], position, rotation mgl32.Vec3, owner scene.Handle) (id scene.Id[*Projectile], proj *Projectile, err error) {
	id, proj, err = st.New()

	proj.owner = owner

	proj.body.Transform = comps.TransformFromTranslationAngles(position, rotation)
	proj.body.Shape = collision.NewSphere(0.7)
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

	proj.moveFunc = func(deltaTime float32) {
		proj.body.Velocity = mgl32.TransformNormal(math2.Vec3Forward(), proj.body.Transform.Matrix())
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
	fmt.Println("Oy blin!")
}
