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

type GrenadeLauncher struct {
	weaponBase
	idleAnim, fireAnim textures.Animation
}

func (grenadeLauncher *GrenadeLauncher) Init(hud *Hud) {
	grenadeLauncher.weaponBase = weaponBase{
		hud:           hud,
		cooldown:      1.0,
		spriteTexture: cache.GetTexture("assets/textures/ui/grenade_launcher_hud.png"),
		swayExtents:   mgl32.Vec2{16.0, 8.0},
		swaySpeed:     mgl32.Vec2{0.75, 1.25},
		ammoType:      game.AMMO_TYPE_GRENADE,
		ammoCost:      1,
	}

	var ok bool
	if grenadeLauncher.idleAnim, ok = grenadeLauncher.spriteTexture.GetAnimation("idle"); !ok {
		log.Println("grenade launcher idle anim not found")
	}
	if grenadeLauncher.fireAnim, ok = grenadeLauncher.spriteTexture.GetAnimation("fire"); !ok {
		log.Println("grenade launcher fire anim not found")
	}
	grenadeLauncher.defaultAnimation = grenadeLauncher.idleAnim

	grenadeLauncher.spriteSize = mgl32.Vec2{
		grenadeLauncher.idleAnim.Frames[0].Rect.Width * SpriteScale(),
		grenadeLauncher.idleAnim.Frames[0].Rect.Height * SpriteScale(),
	}
	grenadeLauncher.spriteEndPos = mgl32.Vec2{
		settings.UIWidth()/2 - grenadeLauncher.spriteSize.X()/2.0,
		settings.UIHeight() - grenadeLauncher.spriteSize.Y(),
	}
	grenadeLauncher.spriteStartPos = grenadeLauncher.spriteEndPos.Add(mgl32.Vec2{0.0, grenadeLauncher.spriteSize.Y()})
}

func (grenadeLauncher *GrenadeLauncher) Order() WeaponIndex {
	return WEAPON_ORDER_GRENADE
}

func (grenadeLauncher *GrenadeLauncher) NoiseLevel() float32 {
	return 50.0
}

func (grenadeLauncher *GrenadeLauncher) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	grenadeLauncher.weaponBase.Update(deltaTime, swayAmount, ammo)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = grenadeLauncher.sprite.Get(); !ok {
		return
	}

	if grenadeLauncher.CanFire(ammo) || ammo[grenadeLauncher.AmmoType()] == 0 {
		sprite.AnimPlayer.ChangeAnimation(grenadeLauncher.idleAnim)
	}
	sprite.AnimPlayer.Update(deltaTime)
}

func (grenadeLauncher *GrenadeLauncher) Fire(ammo *game.Ammo) {
	grenadeLauncher.weaponBase.Fire(ammo)
	if spriteBox, ok := grenadeLauncher.sprite.Get(); ok {
		if spriteBox.AnimPlayer.CurrentAnimation().Name != grenadeLauncher.fireAnim.Name {
			spriteBox.AnimPlayer.ChangeAnimation(grenadeLauncher.fireAnim)
			spriteBox.AnimPlayer.PlayFromStart()
		}
	}
}

func (grenadeLauncher *GrenadeLauncher) IsShooter() bool {
	return true
}
