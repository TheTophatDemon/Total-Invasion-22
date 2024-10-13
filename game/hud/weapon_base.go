package hud

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
)

type WeaponIndex int8

const (
	WEAPON_ORDER_NONE WeaponIndex = iota - 1
	WEAPON_ORDER_SICKLE
	WEAPON_ORDER_CHICKEN
	WEAPON_ORDER_GRENADE
	WEAPON_ORDER_PARUSU
	WEAPON_ORDER_DBL_GRENADE
	WEAPON_ORDER_SIGN
	WEAPON_ORDER_AIRHORN
	WEAPON_ORDER_COUNT
)

const (
	TRANSITION_MOVE_SPEED = 3072.0
)

type weaponState uint8

const (
	WEAPON_STATE_INACTIVE weaponState = iota
	WEAPON_STATE_INTRO
	WEAPON_STATE_READY
	WEAPON_STATE_OUTRO
	WEAPON_STATE_COUNT
)

type weaponBase struct {
	sprite           scene.Id[*ui.Box]
	equipped         bool
	hud              *Hud
	cooldown         float32
	cooldownTimer    float32
	sway             float32     // Value tracking the timeline of the sway animation.
	swayExtents      mgl32.Vec2  // Defines a rectangle on screen within which the weapon will sway
	swaySpeed        mgl32.Vec2  // Defines the speed at which the weapon will sway in each axis
	spriteEndPos     mgl32.Vec2  // Where the weapon sprite's main location is.
	spriteStartPos   mgl32.Vec2  // Position that the sprite moves from when transitioning from a different weapon.
	spriteSize       mgl32.Vec2  // How big the weapon's sprite is on the HUD.
	state            weaponState // Describes the state of transitional animations.
	spriteTexture    *textures.Texture
	defaultAnimation textures.Animation // The animation that plays after the weapon is selected.
	ammoType         game.AmmoType
	ammoCost         int // Amount subtracted from ammo after firing
}

func (wb *weaponBase) Equip() {
	wb.equipped = true
}

func (wb *weaponBase) IsEquipped() bool {
	return wb.equipped
}

func (wb *weaponBase) Select() {
	wb.sway = 0.0
	wb.state = WEAPON_STATE_INTRO
	var (
		spriteBox *ui.Box
		err       error
	)
	wb.sprite, spriteBox, err = wb.hud.UI.Boxes.New()
	if err != nil {
		failure.LogErrWithLocation("%v", err)
		return
	}
	spriteBox.
		SetDest(math2.Rect{
			X:      wb.spriteStartPos.X(),
			Y:      wb.spriteStartPos.Y(),
			Width:  wb.spriteSize.X(),
			Height: wb.spriteSize.Y(),
		}).
		SetTexture(wb.spriteTexture).
		SetColor(color.White).
		SetDepth(0.0)
	spriteBox.AnimPlayer.ChangeAnimation(wb.defaultAnimation)
	spriteBox.AnimPlayer.PlayFromStart()
}

func (wb *weaponBase) Deselect() {
	if wb.state != WEAPON_STATE_INACTIVE {
		wb.state = WEAPON_STATE_OUTRO
	}
}

func (wb *weaponBase) IsSelected() bool {
	return wb.state != WEAPON_STATE_INACTIVE
}

func (wb *weaponBase) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	wb.cooldownTimer = max(wb.cooldownTimer-deltaTime, 0.0)
	swayX := math2.Cos(wb.sway*wb.swaySpeed.X()) * wb.swayExtents.X()
	swayY := math2.Sin(wb.sway*wb.swaySpeed.Y()) * wb.swayExtents.Y()
	if sprite, ok := wb.sprite.Get(); ok {
		switch wb.state {
		case WEAPON_STATE_READY:
			wb.sway += deltaTime * swayAmount
			// Sway the weapon according to player movement
			sprite.SetDestPosition(mgl32.Vec2{
				wb.spriteEndPos.X() + swayX,
				wb.spriteEndPos.Y() + swayY,
			})
		case WEAPON_STATE_INTRO, WEAPON_STATE_OUTRO:
			// Move the weapon towards its screen position.
			target := mgl32.Vec2{swayX, swayY}
			if wb.state == WEAPON_STATE_OUTRO {
				target = target.Add(wb.spriteStartPos)
			} else {
				target = target.Add(wb.spriteEndPos)
			}
			diff := mgl32.Vec2{
				target.X() - sprite.Dest().X,
				target.Y() - sprite.Dest().Y,
			}
			dist := diff.Len()
			moveAmt := deltaTime * TRANSITION_MOVE_SPEED
			if dist < moveAmt {
				sprite.SetDest(math2.Rect{
					X:      target.X(),
					Y:      target.Y(),
					Width:  wb.spriteSize.X(),
					Height: wb.spriteSize.Y(),
				})
				wb.state = (wb.state + 1) % WEAPON_STATE_COUNT
			} else {
				sprite.SetDestPosition(sprite.DestPosition().Add(diff.Mul(moveAmt / dist)))
			}
		}
	}
}

func (wb *weaponBase) Fire(ammo *game.Ammo) {
	wb.cooldownTimer = wb.cooldown
	ammo[wb.ammoType] -= wb.ammoCost
}

func (wb *weaponBase) CanFire(ammo *game.Ammo) bool {
	return wb.state == WEAPON_STATE_READY && wb.cooldownTimer <= 0.0 && ammo[wb.ammoType] >= wb.ammoCost
}

func (wb *weaponBase) AmmoType() game.AmmoType {
	return wb.ammoType
}
