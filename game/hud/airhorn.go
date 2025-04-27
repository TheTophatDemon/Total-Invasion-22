package hud

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	SFX_AIRHORN = "assets/sounds/weapon/airhorn.wav"
)

type Airhorn struct {
	weaponBase
	idleAnim, honkAnim textures.Animation
	heldDown           bool
	honker             tdaudio.VoiceId
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
		settings.UIWidth()/2 - airhorn.spriteSize.X()/2.0 + settings.UIWidth()/6.0,
		settings.UIHeight() - airhorn.spriteSize.Y(),
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

	if !airhorn.heldDown {
		sprite.AnimPlayer.PlayNewAnim(airhorn.idleAnim)
		if airhorn.honker.IsPlaying() && airhorn.honker.GetTime() < 800 {
			airhorn.honker.Seek(800)
		}
	}

	sprite.AnimPlayer.Update(deltaTime)
	airhorn.heldDown = false
}

func (airhorn *Airhorn) Fire(ammo *game.Ammo) {
	airhorn.weaponBase.Fire(ammo)
	airhorn.heldDown = true
	if box, ok := airhorn.sprite.Get(); ok {
		if !box.AnimPlayer.IsPlayingAnim(airhorn.honkAnim) {
			box.AnimPlayer.PlayNewAnim(airhorn.honkAnim)
			airhorn.honker = cache.GetSfx(SFX_AIRHORN).Play()
		}
	}
}

func (airhorn *Airhorn) NoiseLevel() float32 {
	return 100.0
}

func (airhorn *Airhorn) IsShooter() bool {
	return false
}
