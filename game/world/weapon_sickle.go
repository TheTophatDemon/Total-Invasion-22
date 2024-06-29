package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	SFX_SICKLE_CATCH = "assets/sounds/sickle_return.wav"
)

type WeaponSickle struct {
	weaponBase
	throwAnim, catchAnim, idleAnim textures.Animation
	thrownSickle                   scene.Id[*Projectile]
}

var _ Weapon = (*WeaponSickle)(nil)

func NewSickle(world *World, owner scene.Id[HasActor]) *WeaponSickle {
	sickle := &WeaponSickle{
		weaponBase: weaponBase{
			owner:         owner,
			world:         world,
			cooldown:      0.25,
			spriteTexture: cache.GetTexture("assets/textures/ui/sickle_hud.png"),
			swayExtents:   mgl32.Vec2{32.0, 16.0},
			swaySpeed:     mgl32.Vec2{0.75, 1.5},
		},
	}

	var ok bool
	sickle.idleAnim, ok = sickle.spriteTexture.GetAnimation("idle")
	if !ok {
		log.Println("sickle idle anim not found")
	}
	sickle.defaultAnimation = sickle.idleAnim

	sickle.spriteSize = mgl32.Vec2{
		sickle.idleAnim.Frames[0].Rect.Width * 2.0,
		sickle.idleAnim.Frames[0].Rect.Height * 2.0,
	}
	sickle.spriteEndPos = mgl32.Vec2{
		settings.UI_WIDTH/2 - sickle.spriteSize.X()/2.0 + 192.0,
		settings.UI_HEIGHT - sickle.spriteSize.Y() + 16.0,
	}
	sickle.spriteStartPos = sickle.spriteEndPos.Add(mgl32.Vec2{0.0, sickle.spriteSize.Y()})

	sickle.throwAnim, ok = sickle.spriteTexture.GetAnimation("throw")
	if !ok {
		log.Println("sickle throw anim not found")
	}

	sickle.catchAnim, ok = sickle.spriteTexture.GetAnimation("catch")
	if !ok {
		log.Println("sickle catch anim not found")
	}

	return sickle
}

func (sickle *WeaponSickle) Order() WeaponIndex {
	return WEAPON_ORDER_SICKLE
}

func (sickle *WeaponSickle) Update(deltaTime float32, swayAmount float32) {
	sickle.weaponBase.Update(deltaTime, swayAmount)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = sickle.sprite.Get(); !ok {
		return
	}

	sprite.AnimPlayer.Update(deltaTime)

	if !sickle.thrownSickle.Exists() && sprite.AnimPlayer.CurrentAnimation().Name == sickle.throwAnim.Name {
		sprite.AnimPlayer.ChangeAnimation(sickle.catchAnim)
		sprite.AnimPlayer.PlayFromStart()
		cache.GetSfx(SFX_SICKLE_CATCH).Play()
	}
}

func (sickle *WeaponSickle) CanFire() bool {
	return !sickle.thrownSickle.Exists() && sickle.weaponBase.CanFire()
}

func (sickle *WeaponSickle) Fire() {
	sickle.weaponBase.Fire()
	if ownerActor, ok := sickle.owner.Get(); ok {
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, 0.0, -0.5}, ownerActor.Body().Transform.Matrix())
		sickle.thrownSickle, _, _ = SpawnSickle(sickle.world, &sickle.world.Projectiles, firePos, ownerActor.Body().Transform.Rotation(), sickle.owner.Handle)
		if box, ok := sickle.sprite.Get(); ok {
			box.AnimPlayer.ChangeAnimation(sickle.throwAnim)
			box.AnimPlayer.PlayFromStart()
		}
	}
}
