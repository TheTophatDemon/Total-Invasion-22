package hud

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const TEX_WEAPON_SLOT = "assets/textures/ui/weapon_slot.png"

type WeaponWheel struct {
	slots             [WEAPON_ORDER_DEFENESTRATOR]ui.Box
	icons             [WEAPON_ORDER_DEFENESTRATOR]ui.Box
	cursor, highlight ui.Box
	highlightedWeapon WeaponIndex
}

var weaponWheelIcons = [WEAPON_ORDER_COUNT]string{
	WEAPON_ORDER_SICKLE:      "assets/textures/sprites/sickle.png",
	WEAPON_ORDER_CHICKEN:     "assets/textures/sprites/chicken_cannon.png",
	WEAPON_ORDER_GRENADE:     "assets/textures/sprites/grenade_launcher.png",
	WEAPON_ORDER_PARUSU:      "assets/textures/sprites/parusu.png",
	WEAPON_ORDER_DBL_GRENADE: "assets/textures/sprites/double_grenade_launcher.png",
	WEAPON_ORDER_SIGN:        "assets/textures/sprites/sign_of_madness.png",
	WEAPON_ORDER_AIRHORN:     "assets/textures/sprites/airhorn.png",
}

func NewWeaponWheel(weapons []Weapon) WeaponWheel {
	input.UntrapMouse()
	var wheel WeaponWheel
	slotTexture := cache.GetTexture(TEX_WEAPON_SLOT)
	slotWidth := slotTexture.Rect().Width * SpriteScale()
	slotHeight := slotTexture.Rect().Height * SpriteScale()
	for i := range wheel.slots {
		// All slots start in the center
		wheel.slots[i] = ui.NewBoxFull(math2.Rect{
			X:      (settings.UIWidth() / 2.0) - (slotWidth / 2.0),
			Y:      (settings.UIHeight() / 2.0) - (slotHeight / 2.0),
			Width:  slotWidth,
			Height: slotHeight,
		}, slotTexture, weaponColors[i])

		if i == int(WEAPON_ORDER_SICKLE) {
			wheel.slots[i].Depth = 6.1 // Put sickle above other slots while they are expanding out
		} else {
			wheel.slots[i].Depth = 6.0
		}

		if iconPath := weaponWheelIcons[i]; len(iconPath) > 0 {
			iconTex := cache.GetTexture(iconPath)
			iconWidth := min(iconTex.Rect().Width*SpriteScale(), 40*SpriteScale())
			iconHeight := min(iconTex.Rect().Height*SpriteScale(), 40*SpriteScale())
			wheel.icons[i] = ui.NewBoxFull(math2.Rect{
				X:      wheel.slots[i].DestPosition()[0] + (slotWidth-iconWidth)/2.0,
				Y:      wheel.slots[i].DestPosition()[1] + (slotHeight-iconHeight)/2.0,
				Width:  iconWidth,
				Height: iconHeight,
			}, iconTex, color.White)
			wheel.icons[i].Depth = 7.5

			if weapons[i] == nil || !weapons[i].IsEquipped() {
				// Intentionally overflow the color values to erase detail on the image.
				wheel.icons[i].Color = color.Color{R: 100.0, G: 100.0, B: 100.0, A: 0.2}
			}
		}
	}
	input.SetMousePosition(wheel.slots[WEAPON_ORDER_SICKLE].Dest.Center())
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
	slotWidth := slotTexture.Rect().Width * SpriteScale()
	slotHeight := slotTexture.Rect().Height * SpriteScale()
	majorRadiusSq := math2.Pow(slotWidth/2.0, 2.0)
	minorRadiusSq := math2.Pow(slotHeight/2.0, 2.0)

	for i := range wheel.slots {
		slot := &wheel.slots[i]
		icon := &wheel.icons[i]

		if input.IsActionPressed(settings.ACTION_WEAPON_WHEEL) {
			angle := (float32(i-1) / float32(len(wheel.slots)-1)) * math.Pi * 2.0
			targetPos := mgl32.Vec2{
				(settings.UIWidth() / 2.0) - (slotWidth / 2.0) + math2.Sin(angle)*slotWidth*1.5,
				(settings.UIHeight() / 2.0) - (slotHeight / 2.0) - math2.Cos(angle)*slotHeight*1.5,
			}

			if i != int(WEAPON_ORDER_SICKLE) {
				// Move slot towards target position
				dx := (targetPos[0] - slot.Dest.X) * 0.5
				dy := (targetPos[1] - slot.Dest.Y) * 0.5
				slot.Dest.X += dx
				slot.Dest.Y += dy
				icon.Dest.X += dx
				icon.Dest.Y += dy
			}

			// Test intersection with the ellipse
			cx, cy := slot.Dest.Center()
			if (math2.Pow(mousePos[0]-cx, 2.0)/majorRadiusSq)+(math2.Pow(mousePos[1]-cy, 2.0)/minorRadiusSq) <= 1.0 {
				wheel.highlightedWeapon = WeaponIndex(i)
			}
		}

		queue.Add(slot)
		queue.Add(icon)
	}

	if wheel.highlightedWeapon != WEAPON_ORDER_NONE {
		wheel.highlight.SetDestPosition(wheel.slots[int(wheel.highlightedWeapon)].DestPosition())
		queue.Add(&wheel.highlight)
	}

	wheel.cursor.SetDestPosition(mousePos)
	queue.Add(&wheel.cursor)
}
