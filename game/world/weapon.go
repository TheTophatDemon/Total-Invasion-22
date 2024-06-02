package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	WEAPON_ORDER_NONE int = iota - 1
	WEAPON_ORDER_SICKLE
	WEAPON_ORDER_CHICKEN
	WEAPON_ORDER_GRENADE
	WEAPON_ORDER_PARUSU
	WEAPON_ORDER_DBL_GRENADE
	WEAPON_ORDER_SIGN
	WEAPON_ORDER_AIRHORN
	WEAPON_ORDER_MAX
)

type weaponBase struct {
	owner                  scene.Id[HasActor]
	sprite                 scene.Id[*ui.Box]
	equipped               bool
	world                  *World
	cooldown               float32
	cooldownTimer          float32
	sway                   float32
	swayExtents, swaySpeed mgl32.Vec2
	spriteOffset           mgl32.Vec2
	spriteScale            float32
	spriteTexture          *textures.Texture
	defaultAnimation       textures.Animation
}

func (wb *weaponBase) Equip() {
	wb.equipped = true
}

func (wb *weaponBase) Equipped() bool {
	return wb.equipped
}

func (wb *weaponBase) Select() {
	var (
		spriteBox *ui.Box
		err       error
	)
	wb.sprite, spriteBox, err = wb.world.UI.Boxes.New()
	if err != nil {
		log.Println(err)
		return
	}
	spriteBox.
		SetDest(math2.Rect{
			// Position will be set later during Update()
			Width:  wb.defaultAnimation.Frames[0].Rect.Width * wb.spriteScale,
			Height: wb.defaultAnimation.Frames[0].Rect.Height * wb.spriteScale,
		}).
		SetTexture(wb.spriteTexture).
		SetColor(color.White)
	spriteBox.AnimPlayer.ChangeAnimation(wb.defaultAnimation)
	spriteBox.AnimPlayer.PlayFromStart()
}

func (wb *weaponBase) Update(deltaTime float32, swayAmount float32) {
	wb.cooldownTimer = max(wb.cooldownTimer-deltaTime, 0.0)
	wb.sway += deltaTime * swayAmount
	if sprite, ok := wb.sprite.Get(); ok {
		swayX := math2.Cos(wb.sway*wb.swaySpeed.X()) * wb.swayExtents.X()
		swayY := math2.Sin(wb.sway*wb.swaySpeed.Y()) * wb.swayExtents.Y()
		sprite.SetDestPosition(mgl32.Vec2{
			settings.UI_WIDTH/2 - sprite.Dest().Width/2.0 + wb.spriteOffset.X() + swayX,
			settings.UI_HEIGHT - sprite.Dest().Height + wb.spriteOffset.Y() + swayY,
		})
	}
}

func (wb *weaponBase) Fire() {
	wb.cooldownTimer = wb.cooldown
}

func (wb *weaponBase) CanFire() bool {
	return wb.cooldownTimer <= 0.0
}
