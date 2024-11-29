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

type Parusu struct {
	weaponBase
	idleAnim, fireAnim textures.Animation
}

func (parusu *Parusu) Init(hud *Hud) {
	parusu.weaponBase = weaponBase{
		hud:           hud,
		cooldown:      0.10,
		spriteTexture: cache.GetTexture("assets/textures/ui/parusu_hud.png"),
		swayExtents:   mgl32.Vec2{16.0, 8.0},
		swaySpeed:     mgl32.Vec2{0.5, 1.0},
		ammoType:      game.AMMO_TYPE_PLASMA,
		ammoCost:      1,
	}

	var ok bool
	if parusu.idleAnim, ok = parusu.spriteTexture.GetAnimation("idle"); !ok {
		log.Println("parusu idle anim not found")
	}
	if parusu.fireAnim, ok = parusu.spriteTexture.GetAnimation("fire"); !ok {
		log.Println("parusu fire anim not found")
	}
	parusu.defaultAnimation = parusu.idleAnim

	parusu.spriteSize = mgl32.Vec2{
		parusu.idleAnim.Frames[0].Rect.Width * SpriteScale(),
		parusu.idleAnim.Frames[0].Rect.Height * SpriteScale(),
	}
	parusu.spriteEndPos = mgl32.Vec2{
		settings.UIWidth()/2 - parusu.spriteSize.X()/2.0,
		settings.UIHeight() - parusu.spriteSize.Y() + 16.0,
	}
	parusu.spriteStartPos = parusu.spriteEndPos.Add(mgl32.Vec2{0.0, parusu.spriteSize.Y()})
}

func (parusu *Parusu) Order() WeaponIndex {
	return WEAPON_ORDER_PARUSU
}

func (parusu *Parusu) NoiseLevel() float32 {
	return 40.0
}

func (parusu *Parusu) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	parusu.weaponBase.Update(deltaTime, swayAmount, ammo)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = parusu.sprite.Get(); !ok {
		return
	}

	//TODO: Go to idle animation when no longer firing.
	if ammo[parusu.AmmoType()] == 0 {
		sprite.AnimPlayer.ChangeAnimation(parusu.idleAnim)
	}
	sprite.AnimPlayer.Update(deltaTime)
}

func (parusu *Parusu) Fire(ammo *game.Ammo) {
	parusu.weaponBase.Fire(ammo)
	if spriteBox, ok := parusu.sprite.Get(); ok {
		if spriteBox.AnimPlayer.CurrentAnimation().Name != parusu.fireAnim.Name {
			spriteBox.AnimPlayer.ChangeAnimation(parusu.fireAnim)
			spriteBox.AnimPlayer.PlayFromStart()
		}
	}
}
