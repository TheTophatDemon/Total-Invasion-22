package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type WeaponChicken struct {
	weaponBase
	hudTexture         *textures.Texture
	idleAnim, fireAnim textures.Animation
}

var _ Weapon = (*WeaponChicken)(nil)

func NewChickenCannon(world *World, owner scene.Id[HasActor]) WeaponChicken {
	chicken := WeaponChicken{
		weaponBase: weaponBase{
			owner:    owner,
			world:    world,
			cooldown: 0.1,
		},
	}

	chicken.hudTexture = cache.GetTexture("assets/textures/ui/chicken_cannon_hud.png")

	var ok bool
	if chicken.idleAnim, ok = chicken.hudTexture.GetAnimation("idle"); !ok {
		log.Println("chicken cannon idle anim not found")
	}
	if chicken.fireAnim, ok = chicken.hudTexture.GetAnimation("fire"); !ok {
		log.Println("chicken cannon fire anim not found")
	}

	return chicken
}

func (chicken *WeaponChicken) Order() int {
	return WEAPON_ORDER_CHICKEN
}

func (chicken *WeaponChicken) Equip() {
	chicken.weaponBase.Equip()
}

func (chicken *WeaponChicken) Select() {
	var (
		spriteBox *ui.Box
		err       error
	)
	chicken.sprite, spriteBox, err = chicken.world.UI.Boxes.New()
	if err != nil {
		log.Println(err)
		return
	}
	spriteBox.
		SetSrc(math2.Rect{
			X: 256.0, Y: 0.0,
			Width: 256.0, Height: 192.0,
		}).
		SetDest(math2.Rect{
			X:     float32(settings.UI_WIDTH/2) - 256.0,
			Y:     settings.UI_HEIGHT - 192.0*2.0,
			Width: 512.0, Height: 192.0 * 2.0,
		}).
		SetTexture(chicken.hudTexture).
		SetColor(color.White)

	spriteBox.AnimPlayer.ChangeAnimation(chicken.idleAnim)
	spriteBox.AnimPlayer.PlayFromStart()
}

func (chicken *WeaponChicken) Deselect() {
	chicken.sprite.Remove()
}

func (chicken *WeaponChicken) Update(deltaTime float32) {
	chicken.weaponBase.Update(deltaTime)

	var sprite *ui.Box
	var ok bool
	if sprite, ok = chicken.sprite.Get(); !ok {
		return
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
