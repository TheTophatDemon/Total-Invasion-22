package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/hud"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type WeaponChicken struct {
	weaponBase
	idleAnim, fireAnim textures.Animation
}

var _ Weapon = (*WeaponChicken)(nil)

func NewChickenCannon(world *World, owner scene.Id[HasActor]) *WeaponChicken {
	chickenGun := &WeaponChicken{
		weaponBase: weaponBase{
			owner:         owner,
			world:         world,
			cooldown:      0.15,
			spriteTexture: cache.GetTexture("assets/textures/ui/chicken_cannon_hud.png"),
			swayExtents:   mgl32.Vec2{16.0, 8.0},
			swaySpeed:     mgl32.Vec2{0.5, 1.0},
		},
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
		chickenGun.idleAnim.Frames[0].Rect.Width * hud.SpriteScale(),
		chickenGun.idleAnim.Frames[0].Rect.Height * hud.SpriteScale(),
	}
	chickenGun.spriteEndPos = mgl32.Vec2{
		settings.UIWidth()/2 - chickenGun.spriteSize.X()/2.0,
		settings.UIHeight() - chickenGun.spriteSize.Y() + 16.0,
	}
	chickenGun.spriteStartPos = chickenGun.spriteEndPos.Add(mgl32.Vec2{0.0, chickenGun.spriteSize.Y()})

	return chickenGun
}

func (chickenGun *WeaponChicken) Order() WeaponIndex {
	return WEAPON_ORDER_CHICKEN
}

func (chickenGun *WeaponChicken) AmmoType() AmmoType {
	return AMMO_TYPE_EGG
}

func (chickenGun *WeaponChicken) Update(deltaTime float32, swayAmount float32) {
	chickenGun.weaponBase.Update(deltaTime, swayAmount)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = chickenGun.sprite.Get(); !ok {
		return
	}

	if chickenGun.CanFire() {
		sprite.AnimPlayer.ChangeAnimation(chickenGun.idleAnim)
	}
	sprite.AnimPlayer.Update(deltaTime)
}

func (chickenGun *WeaponChicken) Fire() {
	chickenGun.weaponBase.Fire()
	if ownerActor, ok := chickenGun.owner.Get(); ok {
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, -0.15, -0.5}, ownerActor.Body().Transform.Matrix())
		SpawnEgg(chickenGun.world, &chickenGun.world.Projectiles, firePos, ownerActor.Body().Transform.Rotation(), chickenGun.owner.Handle)
	}

	if spriteBox, ok := chickenGun.sprite.Get(); ok {
		if spriteBox.AnimPlayer.CurrentAnimation().Name != chickenGun.fireAnim.Name {
			spriteBox.AnimPlayer.ChangeAnimation(chickenGun.fireAnim)
			spriteBox.AnimPlayer.PlayFromStart()
		}
	}
}
