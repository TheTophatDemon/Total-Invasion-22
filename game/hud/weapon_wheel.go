package hud

import (
	"math"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const TEX_WEAPON_SLOT = "assets/textures/ui/weapon_slot.png"

type WeaponWheel struct {
	slots             [WEAPON_ORDER_COUNT]ui.Box
	icons             [WEAPON_ORDER_COUNT]ui.Box
	cursor, highlight ui.Box
	highlightedWeapon WeaponIndex
}

var weaponWheelIcons = [WEAPON_ORDER_COUNT]string{
	WEAPON_ORDER_SICKLE:  "assets/textures/sprites/sickle.png",
	WEAPON_ORDER_CHICKEN: "assets/textures/sprites/chicken_cannon.png",
	WEAPON_ORDER_GRENADE: "assets/textures/sprites/grenade_launcher.png",
	WEAPON_ORDER_PARUSU:  "assets/textures/sprites/parusu.png",
	WEAPON_ORDER_AIRHORN: "assets/textures/sprites/airhorn.png",
}

func NewWeaponWheel(currentWeapon WeaponIndex) WeaponWheel {
	input.UntrapMouse()
	var wheel WeaponWheel
	slotTexture := cache.GetTexture(TEX_WEAPON_SLOT)
	slotWidth := slotTexture.Rect().Width * SpriteScale()
	slotHeight := slotTexture.Rect().Height * SpriteScale()
	for i := range wheel.slots {
		if i == int(WEAPON_ORDER_SICKLE) {
			// Sickle goes in the center
			wheel.slots[i] = ui.NewBoxFull(math2.Rect{
				X:      (settings.UIWidth() / 2.0) - (slotWidth / 2.0),
				Y:      (settings.UIHeight() / 2.0) - (slotHeight / 2.0),
				Width:  slotWidth,
				Height: slotHeight,
			}, slotTexture, weaponColors[i])
		} else {
			// Other weapons are in a circle surrounding the sickle
			angle := (float32(i-1) / float32(WEAPON_ORDER_COUNT-1)) * math.Pi * 2.0
			wheel.slots[i] = ui.NewBoxFull(math2.Rect{
				X:      (settings.UIWidth() / 2.0) - (slotWidth / 2.0) + math2.Sin(angle)*slotWidth*1.5,
				Y:      (settings.UIHeight() / 2.0) - (slotHeight / 2.0) - math2.Cos(angle)*slotHeight*1.5,
				Width:  slotWidth,
				Height: slotHeight,
			}, slotTexture, weaponColors[i])
		}
		wheel.slots[i].Depth = 6.0

		if iconPath := weaponWheelIcons[i]; len(iconPath) > 0 {
			iconTex := cache.GetTexture(iconPath)
			iconWidth := iconTex.Rect().Width * SpriteScale()
			iconHeight := iconTex.Rect().Height * SpriteScale()
			wheel.icons[i] = ui.NewBoxFull(math2.Rect{
				X:      wheel.slots[i].DestPosition()[0] + (slotWidth-iconWidth)/2.0,
				Y:      wheel.slots[i].DestPosition()[1] + (slotHeight-iconHeight)/2.0,
				Width:  iconWidth,
				Height: iconHeight,
			}, iconTex, color.White)
			wheel.icons[i].Depth = 7.5
		}

		if i == int(currentWeapon) {
			input.SetMousePosition(wheel.slots[i].Dest.Center())
		}
	}
	cursorTex := cache.GetTexture("assets/textures/ui/hand_cursor.png")
	wheel.cursor = ui.NewBoxFull(
		math2.Rect{
			Width:  cursorTex.Rect().Width * SpriteScale(),
			Height: cursorTex.Rect().Height * SpriteScale(),
		},
		cursorTex,
		color.White)
	wheel.cursor.Depth = 8.0

	wheel.highlight = ui.NewBoxFull(math2.Rect{
		X:      (settings.UIWidth() / 2.0) - (slotWidth / 2.0),
		Y:      (settings.UIHeight() / 2.0) - (slotHeight / 2.0),
		Width:  slotWidth,
		Height: slotHeight,
	}, slotTexture, color.White)
	wheel.highlight.Depth = 7.0

	return wheel
}

func (wheel *WeaponWheel) Render(queue *ui.RenderQueue) {
	mousePos := input.MousePosition()
	slotTexture := cache.GetTexture(TEX_WEAPON_SLOT)
	majorRadiusSq := math2.Pow(slotTexture.Rect().Width*SpriteScale()/2.0, 2.0)
	minorRadiusSq := math2.Pow(slotTexture.Rect().Height*SpriteScale()/2.0, 2.0)

	for i := range wheel.slots {
		// Test intersection with the ellipse
		cx, cy := wheel.slots[i].Dest.Center()
		if (math2.Pow(mousePos[0]-cx, 2.0)/majorRadiusSq)+(math2.Pow(mousePos[1]-cy, 2.0)/minorRadiusSq) <= 1.0 {
			wheel.highlightedWeapon = WeaponIndex(i)
		}

		queue.Add(&wheel.slots[i])
		queue.Add(&wheel.icons[i])
	}

	if wheel.highlightedWeapon != WEAPON_ORDER_NONE {
		wheel.highlight.SetDestPosition(wheel.slots[int(wheel.highlightedWeapon)].DestPosition())
		queue.Add(&wheel.highlight)
	}

	wheel.cursor.SetDestPosition(mousePos)
	queue.Add(&wheel.cursor)
}
