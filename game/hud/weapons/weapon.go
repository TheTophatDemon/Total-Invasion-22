package weapons

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type WeaponKind int8

const (
	WeaponNone WeaponKind = iota
	WeaponSickle
	WeaponChicken
	WeaponGrenade
	WeaponParusu
	WeaponDblGrenade
	WeaponSign
	WeaponAirhorn
	WeaponDefenestrator
	WeaponCluckster
	WeaponCount
)

type weaponState uint8

const (
	weaponStateInactive weaponState = iota
	weaponStateIntro
	weaponStateReady
	weaponStateOutro
	weaponStateCount
)

type Weapon struct {
	kind          WeaponKind
	name          string
	texturePath   string
	sprite        ui.Box
	equipped      bool
	cooldown      float32
	cooldownTimer float32
	sway          float32     // Value tracking the timeline of the sway animation.
	swayExtents   mgl32.Vec2  // Defines a rectangle on screen within which the weapon will sway
	swaySpeed     mgl32.Vec2  // Defines the speed at which the weapon will sway in each axis
	spriteOffset  mgl32.Vec2  // Offset from the bottom center of the screen where the sprite will rest
	state         weaponState // Describes the state of transitional animations.
	ammoType      game.AmmoType
	ammoCost      int         // Amount subtracted from ammo after firing
	wheelColor    color.Color // Color of weapon wheel frame
	wheelIconPath string      // Path to icon texture displayed in weapon wheel slot
	noiseLevel    float32
	isShooter     bool
	updateFunc    func(w *Weapon, deltaTime float32, ammo *game.Ammo)
}

func (wb *Weapon) Equip() {
	wb.equipped = true
}

func (wb *Weapon) IsEquipped() bool {
	return wb.equipped
}

func (wb *Weapon) Select() {
	wb.sway = 0.0
	wb.state = weaponStateIntro

	texture := cache.GetTexture(wb.texturePath)
	idleAnim, ok := texture.GetAnimation("idle")
	if !ok {
		log.Printf("weapon %v missing idle animation\n", wb.name)
	}

	spriteSize := mgl32.Vec2{
		idleAnim.Frames[0].Rect.Width * settings.UIScale() * 2.0,
		idleAnim.Frames[0].Rect.Height * settings.UIScale() * 2.0,
	}
	wb.sprite = ui.Box{
		Color: color.White,
		Src:   math2.Rect{Width: 1.0, Height: 1.0},
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      wb.spriteStartPos.X(),
				Y:      wb.spriteStartPos.Y(),
				Width:  spriteSize[0],
				Height: spriteSize[1],
			},
		},
	}
	wb.sprite.AnimPlayer.ChangeAnimation(idleAnim)
	wb.sprite.AnimPlayer.PlayFromStart()
}

func (wb *Weapon) Deselect() {
	if wb.state != weaponStateInactive {
		wb.state = weaponStateOutro
	}
}

func (wb *Weapon) IsSelected() bool {
	return wb.state != weaponStateInactive
}

func (wb *Weapon) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	wb.sprite.AnimPlayer.Update(deltaTime)
	wb.cooldownTimer = max(wb.cooldownTimer-deltaTime, 0.0)
	swayX := math2.Cos(wb.sway*wb.swaySpeed.X()) * wb.swayExtents.X()
	swayY := math2.Sin(wb.sway*wb.swaySpeed.Y()) * wb.swayExtents.Y()
	switch wb.state {
	case weaponStateReady:
		wb.sway += deltaTime * swayAmount
		// Sway the weapon according to player movement
		wb.sprite.SetDestPosition(mgl32.Vec2{
			wb.spriteEndPos.X() + swayX,
			wb.spriteEndPos.Y() + swayY,
		})
	case weaponStateIntro, weaponStateOutro:
		// Move the weapon towards its screen position.
		target := mgl32.Vec2{swayX, swayY}
		if wb.state == weaponStateOutro {
			target = target.Add(wb.spriteStartPos)
		} else {
			target = target.Add(wb.spriteEndPos)
		}
		diff := mgl32.Vec2{
			target[0] - wb.sprite.Dest.X,
			target[1] - wb.sprite.Dest.Y,
		}
		dist := diff.Len()
		moveAmt := deltaTime * 3072.0
		if dist < moveAmt {
			wb.sprite.SetDestPosition(target)
			wb.state = (wb.state + 1) % weaponStateCount
		} else {
			wb.sprite.SetDestPosition(wb.sprite.DestPosition().Add(diff.Mul(moveAmt / dist)))
		}
	}

	if wb.updateFunc != nil {
		wb.updateFunc(wb, deltaTime, ammo)
	}
}

func (wb *Weapon) Layout(queue *ui.RenderQueue) {
	queue.Add(&wb.sprite)
}

func (wb *Weapon) Fire(ammo *game.Ammo) {
	wb.cooldownTimer = wb.cooldown
	ammo[wb.ammoType] -= wb.ammoCost

	fireAnim, ok := wb.sprite.Texture.GetAnimation("fire")
	if !ok {
		log.Printf("fire anim not found for weapon %v\n", wb.name)
		return
	}

	if wb.sprite.AnimPlayer.CurrentAnimation().Name != fireAnim.Name {
		wb.sprite.AnimPlayer.ChangeAnimation(fireAnim)
		wb.sprite.AnimPlayer.PlayFromStart()
	}
}

func (wb *Weapon) CanFire(ammo *game.Ammo) bool {
	return wb.state == weaponStateReady && wb.cooldownTimer <= 0.0 && ammo[wb.ammoType] >= wb.ammoCost
}

func (wb *Weapon) AmmoType() game.AmmoType {
	return wb.ammoType
}
