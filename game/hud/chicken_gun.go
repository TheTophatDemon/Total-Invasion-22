package hud

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type ChickenGun struct {
	weaponBase
	idleAnim, fireAnim textures.Animation
}

func (chickenGun *ChickenGun) Init(hud *Hud) {
	chickenGun.weaponBase = weaponBase{
		hud:           hud,
		cooldown:      0.15,
		spriteTexture: cache.GetTexture("assets/textures/ui/chicken_cannon_hud.png"),
		swayExtents:   mgl32.Vec2{16.0, 8.0},
		swaySpeed:     mgl32.Vec2{0.5, 1.0},
		ammoType:      game.AMMO_TYPE_EGG,
		ammoCost:      1,
	}

	var ok bool
	if chickenGun.idleAnim, ok = chickenGun.spriteTexture.GetAnimation("idle"); !ok {
		log.Println("chicken cannon idle anim not found")
	}
	if chickenGun.fireAnim, ok = chickenGun.spriteTexture.GetAnimation("fire"); !ok {
		log.Println("chicken cannon fire anim not found")
	}
	chickenGun.defaultAnimation = chickenGun.idleAnim

	chickenGun.spriteSize = mgl32.Vec2{
		chickenGun.idleAnim.Frames[0].Rect.Width * SpriteScale(),
		chickenGun.idleAnim.Frames[0].Rect.Height * SpriteScale(),
	}
	chickenGun.spriteEndPos = mgl32.Vec2{
		settings.UIWidth()/2 - chickenGun.spriteSize.X()/2.0,
		settings.UIHeight() - chickenGun.spriteSize.Y() + 16.0,
	}
	chickenGun.spriteStartPos = chickenGun.spriteEndPos.Add(mgl32.Vec2{0.0, chickenGun.spriteSize.Y()})
}

func (chickenGun *ChickenGun) Order() WeaponIndex {
	return WEAPON_ORDER_CHICKEN
}

func (chickenGun *ChickenGun) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	chickenGun.weaponBase.Update(deltaTime, swayAmount, ammo)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = chickenGun.sprite.Get(); !ok {
		return
	}

	if chickenGun.CanFire(ammo) || ammo[chickenGun.AmmoType()] == 0 {
		sprite.AnimPlayer.ChangeAnimation(chickenGun.idleAnim)
	}
	sprite.AnimPlayer.Update(deltaTime)
}

func (chickenGun *ChickenGun) Fire(ammo *game.Ammo) {
	chickenGun.weaponBase.Fire(ammo)
	if spriteBox, ok := chickenGun.sprite.Get(); ok {
		if spriteBox.AnimPlayer.CurrentAnimation().Name != chickenGun.fireAnim.Name {
			spriteBox.AnimPlayer.ChangeAnimation(chickenGun.fireAnim)
			spriteBox.AnimPlayer.PlayFromStart()
		}
	}
}
