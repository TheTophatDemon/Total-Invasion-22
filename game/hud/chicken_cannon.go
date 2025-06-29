package hud

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type ChickenCannon struct {
	weaponBase
	idleAnim, fireAnim textures.Animation
}

func (chickenCannon *ChickenCannon) Init() {
	chickenCannon.weaponBase = weaponBase{
		cooldown:      0.15,
		spriteTexture: cache.GetTexture("assets/textures/ui/chicken_cannon_hud.png"),
		swayExtents:   mgl32.Vec2{16.0, 8.0},
		swaySpeed:     mgl32.Vec2{0.5, 1.0},
		ammoType:      game.AMMO_TYPE_EGG,
		ammoCost:      1,
	}

	var ok bool
	if chickenCannon.idleAnim, ok = chickenCannon.spriteTexture.GetAnimation("idle"); !ok {
		log.Println("chicken cannon idle anim not found")
	}

	chickenCannon.defaultAnimation = chickenCannon.idleAnim

	chickenCannon.spriteSize = mgl32.Vec2{
		chickenCannon.idleAnim.Frames[0].Rect.Width * SpriteScale(),
		chickenCannon.idleAnim.Frames[0].Rect.Height * SpriteScale(),
	}
	chickenCannon.spriteEndPos = mgl32.Vec2{
		settings.UIWidth()/2 - chickenCannon.spriteSize.X()/2.0,
		settings.UIHeight() - chickenCannon.spriteSize.Y(),
	}
	chickenCannon.spriteStartPos = chickenCannon.spriteEndPos.Add(mgl32.Vec2{0.0, chickenCannon.spriteSize.Y()})
}

func (chickenCannon *ChickenCannon) Order() WeaponIndex {
	return WEAPON_ORDER_CHICKEN
}

func (chickenCannon *ChickenCannon) NoiseLevel() float32 {
	return 20.0
}

func (chickenCannon *ChickenCannon) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	chickenCannon.weaponBase.Update(deltaTime, swayAmount, ammo)

	if chickenCannon.CanFire(ammo) || ammo[chickenCannon.AmmoType()] == 0 {
		chickenCannon.sprite.AnimPlayer.ChangeAnimation(chickenCannon.idleAnim)
	}
	chickenCannon.sprite.AnimPlayer.Update(deltaTime)
}

func (chickenCannon *ChickenCannon) Fire(ammo *game.Ammo) {
	chickenCannon.weaponBase.Fire(ammo)

	fireAnim, ok := chickenCannon.spriteTexture.GetAnimation("fire")
	if !ok {
		log.Println("chicken cannon fire anim not found")
		return
	}
	if chickenCannon.sprite.AnimPlayer.CurrentAnimation().Name != fireAnim.Name {
		chickenCannon.sprite.AnimPlayer.ChangeAnimation(chickenCannon.fireAnim)
		chickenCannon.sprite.AnimPlayer.PlayFromStart()
	}
}

func (chickenCannon *ChickenCannon) IsShooter() bool {
	return true
}

func (chickenCannon *ChickenCannon) WheelColor() color.Color {
	return color.FromBytes(0, 0, 255, 255)
}

func (chickenCannon *ChickenCannon) WheelIconPath() string {
	return "assets/textures/sprites/chicken_cannon.png"
}
