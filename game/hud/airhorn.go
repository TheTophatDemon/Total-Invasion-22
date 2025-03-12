package hud

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Airhorn struct {
	weaponBase
	idleAnim, honkAnim textures.Animation
}

func (airhorn *Airhorn) Init(hud *Hud) {
	airhorn.weaponBase = weaponBase{
		hud:           hud,
		cooldown:      0.0,
		spriteTexture: cache.GetTexture("assets/textures/ui/airhorn_hud.png"),
		swayExtents:   mgl32.Vec2{32.0, 16.0},
		swaySpeed:     mgl32.Vec2{0.75, 1.5},
		ammoType:      game.AMMO_TYPE_NONE,
	}

	var ok bool
	airhorn.idleAnim, ok = airhorn.spriteTexture.GetAnimation("idle")
	if !ok {
		log.Println("airhorn idle anim not found")
	}
	airhorn.defaultAnimation = airhorn.idleAnim

	airhorn.spriteSize = mgl32.Vec2{
		airhorn.idleAnim.Frames[0].Rect.Width * SpriteScale(),
		airhorn.idleAnim.Frames[0].Rect.Height * SpriteScale(),
	}
	airhorn.spriteEndPos = mgl32.Vec2{
		settings.UIWidth()/2 - airhorn.spriteSize.X()/2.0 + settings.UIWidth()/4.0,
		settings.UIHeight() - airhorn.spriteSize.Y() + 16.0,
	}
	airhorn.spriteStartPos = airhorn.spriteEndPos.Add(mgl32.Vec2{0.0, airhorn.spriteSize.Y()})

	airhorn.honkAnim, ok = airhorn.spriteTexture.GetAnimation("honk")
	if !ok {
		log.Println("airhorn honk anim not found")
	}
}

func (airhorn *Airhorn) Order() WeaponIndex {
	return WEAPON_ORDER_AIRHORN
}

func (airhorn *Airhorn) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	airhorn.weaponBase.Update(deltaTime, swayAmount, ammo)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = airhorn.sprite.Get(); !ok {
		return
	}

	if sprite.AnimPlayer.IsPlayingAnim(airhorn.honkAnim) && sprite.AnimPlayer.IsAtEnd() && !input.IsActionPressed(settings.ACTION_FIRE) {
		sprite.AnimPlayer.PlayNewAnim(airhorn.idleAnim)
	}

	sprite.AnimPlayer.Update(deltaTime)
}

func (airhorn *Airhorn) Fire(ammo *game.Ammo) {
	airhorn.weaponBase.Fire(ammo)
	if box, ok := airhorn.sprite.Get(); ok {
		box.AnimPlayer.ChangeAnimation(airhorn.honkAnim)
		if !box.AnimPlayer.IsPlaying() {
			box.AnimPlayer.PlayFromStart()
		}
	}
}

func (airhorn *Airhorn) NoiseLevel() float32 {
	return 100.0
}
