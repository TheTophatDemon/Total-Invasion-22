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

const (
	SFX_SICKLE_CATCH = "assets/sounds/weapon/sickle_return.wav"
)

type Sickle struct {
	weaponBase
	throwAnim, catchAnim, idleAnim textures.Animation
}

func (sickle *Sickle) Init(hud *Hud) {
	sickle.weaponBase = weaponBase{
		hud:           hud,
		cooldown:      0.25,
		spriteTexture: cache.GetTexture("assets/textures/ui/sickle_hud.png"),
		swayExtents:   mgl32.Vec2{32.0, 16.0},
		swaySpeed:     mgl32.Vec2{0.75, 1.5},
		ammoType:      game.AMMO_TYPE_SICKLE,
		ammoCost:      1,
	}

	var ok bool
	sickle.throwAnim, ok = sickle.spriteTexture.GetAnimation("throw")
	if !ok {
		log.Println("sickle throw anim not found")
	}

	sickle.idleAnim, ok = sickle.spriteTexture.GetAnimation("idle")
	if !ok {
		log.Println("sickle idle anim not found")
	}

	sickle.spriteSize = mgl32.Vec2{
		sickle.idleAnim.Frames[0].Rect.Width * SpriteScale(),
		sickle.idleAnim.Frames[0].Rect.Height * SpriteScale(),
	}
	sickle.spriteEndPos = mgl32.Vec2{
		settings.UIWidth()/2 - sickle.spriteSize.X()/2.0 + settings.UIWidth()/4.0,
		settings.UIHeight() - sickle.spriteSize.Y(),
	}
	sickle.spriteStartPos = sickle.spriteEndPos.Add(mgl32.Vec2{0.0, sickle.spriteSize.Y()})

	sickle.catchAnim, ok = sickle.spriteTexture.GetAnimation("catch")
	if !ok {
		log.Println("sickle catch anim not found")
	}
	// We want to start with the throw animation for the level intro and then use the catch animation later.
	sickle.defaultAnimation = sickle.throwAnim
}

func (sickle *Sickle) Select() {
	sickle.weaponBase.Select()
	// Reset the animation that plays on selection after the level intro.
	sickle.defaultAnimation = sickle.catchAnim
}

func (sickle *Sickle) Order() WeaponIndex {
	return WEAPON_ORDER_SICKLE
}

func (sickle *Sickle) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	sickle.weaponBase.Update(deltaTime, swayAmount, ammo)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = sickle.sprite.Get(); !ok {
		return
	}

	sprite.AnimPlayer.Update(deltaTime)

	if sickle.CanFire(ammo) && sprite.AnimPlayer.CurrentAnimation().Name == sickle.throwAnim.Name {
		sprite.AnimPlayer.ChangeAnimation(sickle.catchAnim)
		sprite.AnimPlayer.PlayFromStart()
		cache.GetSfx(SFX_SICKLE_CATCH).Play()
	}
}

func (sickle *Sickle) Fire(ammo *game.Ammo) {
	sickle.weaponBase.Fire(ammo)
	if box, ok := sickle.sprite.Get(); ok {
		box.AnimPlayer.ChangeAnimation(sickle.throwAnim)
		box.AnimPlayer.PlayFromStart()
	}
}

func (sickle *Sickle) NoiseLevel() float32 {
	return 10.0
}

func (sickle *Sickle) IsShooter() bool {
	return true
}
