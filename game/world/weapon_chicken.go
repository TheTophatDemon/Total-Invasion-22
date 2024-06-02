package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

type WeaponChicken struct {
	weaponBase
	idleAnim, fireAnim textures.Animation
}

var _ Weapon = (*WeaponChicken)(nil)

func NewChickenCannon(world *World, owner scene.Id[HasActor]) WeaponChicken {
	chicken := WeaponChicken{
		weaponBase: weaponBase{
			owner:         owner,
			world:         world,
			cooldown:      0.15,
			spriteScale:   2.0,
			spriteTexture: cache.GetTexture("assets/textures/ui/chicken_cannon_hud.png"),
			spriteOffset:  mgl32.Vec2{0.0, 16.0},
			swayExtents:   mgl32.Vec2{16.0, 8.0},
			swaySpeed:     mgl32.Vec2{0.5, 1.0},
		},
	}

	var ok bool
	if chicken.idleAnim, ok = chicken.spriteTexture.GetAnimation("idle"); !ok {
		log.Println("chicken cannon idle anim not found")
	}
	if chicken.fireAnim, ok = chicken.spriteTexture.GetAnimation("fire"); !ok {
		log.Println("chicken cannon fire anim not found")
	}
	chicken.defaultAnimation = chicken.idleAnim

	return chicken
}

func (chicken *WeaponChicken) Order() int {
	return WEAPON_ORDER_CHICKEN
}

func (chicken *WeaponChicken) Equip() {
	chicken.weaponBase.Equip()
}

func (chicken *WeaponChicken) Select() {
	chicken.weaponBase.Select()
}

func (chicken *WeaponChicken) Deselect() {
	chicken.sprite.Remove()
}

func (chicken *WeaponChicken) Update(deltaTime float32, swayAmount float32) {
	chicken.weaponBase.Update(deltaTime, swayAmount)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = chicken.sprite.Get(); !ok {
		return
	}

	if chicken.CanFire() {
		sprite.AnimPlayer.ChangeAnimation(chicken.idleAnim)
	}
	sprite.AnimPlayer.Update(deltaTime)
}

func (chicken *WeaponChicken) Fire() {
	chicken.weaponBase.Fire()
	if ownerActor, ok := chicken.owner.Get(); ok {
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, -0.15, -0.5}, ownerActor.Body().Transform.Matrix())
		SpawnEgg(chicken.world, &chicken.world.Projectiles, firePos, ownerActor.Body().Transform.Rotation(), chicken.owner.Handle)
	}

	if spriteBox, ok := chicken.sprite.Get(); ok {
		if spriteBox.AnimPlayer.CurrentAnimation().Name != chicken.fireAnim.Name {
			spriteBox.AnimPlayer.ChangeAnimation(chicken.fireAnim)
			spriteBox.AnimPlayer.PlayFromStart()
		}
	}
}
