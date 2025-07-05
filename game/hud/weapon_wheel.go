package hud

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const texWeaponSlot = "assets/textures/ui/weapon_slot.png"
const wheelIconTransparency = 0.2

type weaponSlot struct {
	kind      WeaponKind
	back      ui.Box
	icon      ui.Box
	targetPos mgl32.Vec2 // The position where each slot will move towards
}

type WeaponWheel struct {
	slots             [3][3]weaponSlot
	highlight         ui.Box
	highlightedWeapon WeaponKind
	selectPos         mgl32.Vec2 // Represents a virtual mouse position for selecting the weapon
	bounds            math2.Rect // Rectangular region within which the wheel resides
	cursor            ui.Box
}

func newWeaponWheel(weapons []Weapon) WeaponWheel {
	const slotMargin = 16.0
	slotTexture := cache.GetTexture(texWeaponSlot)
	slotWidth := slotTexture.Rect().Width * settings.SpriteScale()
	slotHeight := slotTexture.Rect().Height * settings.SpriteScale()

	wheel := WeaponWheel{
		selectPos: mgl32.Vec2{settings.UIWidth() / 2.0, settings.UIHeight() / 2.0},
		bounds: math2.Rect{
			X:      (settings.UIWidth() / 2.0) - slotMargin - (slotWidth * 1.5),
			Y:      (settings.UIHeight() / 2.0) - slotMargin - (slotHeight * 1.5),
			Width:  (slotWidth * 3.0) + (slotMargin * 2.0),
			Height: (slotHeight * 3.0) + (slotMargin * 2.0),
		},
		highlightedWeapon: WeaponSickle,
	}

	slotStart := wheel.selectPos.Sub(mgl32.Vec2{slotWidth / 2.0, slotHeight / 2.0})

	for i, kind := range [...]WeaponKind{
		WeaponCluckster, WeaponChicken, WeaponAirhorn,
		WeaponSign, WeaponSickle, WeaponGrenade,
		WeaponDefenestrator, WeaponParusu, WeaponDblGrenade,
	} {
		weapon := weapons[kind]

		depth := float32(6.0)
		if kind == WeaponSickle {
			depth = 6.1 // Put sickle above other slots while they are expanding out
		}

		var icon ui.Box
		if iconPath := weapon.wheelIconPath; len(iconPath) > 0 {
			iconTex := cache.GetTexture(iconPath)
			iconWidth := min(iconTex.Rect().Width*settings.SpriteScale(), 40*settings.SpriteScale())
			iconHeight := min(iconTex.Rect().Height*settings.SpriteScale(), 40*settings.SpriteScale())

			icon = ui.NewBoxFull(math2.Rect{
				X:      slotStart[0] + (slotWidth-iconWidth)/2.0,
				Y:      slotStart[1] + (slotHeight-iconHeight)/2.0,
				Width:  iconWidth,
				Height: iconHeight,
			}, iconTex, color.White, 7.5)

			if !weapon.Equipped {
				// Intentionally overflow the color values to erase detail on the image.
				icon.Color = color.Color{R: 100.0, G: 100.0, B: 100.0, A: wheelIconTransparency}
			}
		}

		x, y := i%len(wheel.slots[0]), i/len(wheel.slots[0])

		slot := weaponSlot{
			kind: kind,
			back: ui.NewBoxFull(math2.Rect{
				// All slots start in the center and move towards their target positions
				X:      slotStart[0],
				Y:      slotStart[1],
				Width:  slotWidth,
				Height: slotHeight,
			}, slotTexture, weapon.wheelColor, depth),
			icon: icon,
			targetPos: mgl32.Vec2{
				wheel.bounds.X + float32(x)*(slotWidth+slotMargin),
				wheel.bounds.Y + float32(y)*(slotHeight+slotMargin),
			},
		}

		wheel.slots[x][y] = slot
	}

	cursorTex := cache.GetTexture("assets/textures/ui/hand_cursor.png")
	wheel.cursor = ui.NewBoxFull(
		math2.Rect{
			Width:  cursorTex.Rect().Width * settings.SpriteScale(),
			Height: cursorTex.Rect().Height * settings.SpriteScale(),
		},
		cursorTex,
		color.White,
		8.0,
	)

	wheel.highlight = ui.NewBoxFull(math2.Rect{
		X:      (settings.UIWidth() / 2.0) - (slotWidth / 2.0),
		Y:      (settings.UIHeight() / 2.0) - (slotHeight / 2.0),
		Width:  slotWidth,
		Height: slotHeight,
	}, slotTexture, color.White, 7.0)

	return wheel
}

func (wheel *WeaponWheel) Layout(queue *ui.RenderQueue, openness float32) {
	wheel.selectPos = wheel.selectPos.Add(mgl32.Vec2{
		input.ActionAxis(settings.ACTION_LOOK_HORZ) / settings.Current.MouseSensitivity,
		input.ActionAxis(settings.ACTION_LOOK_VERT) / settings.Current.MouseSensitivity,
	})

	wheel.selectPos[0] = math2.Clamp(wheel.selectPos[0], wheel.bounds.X, wheel.bounds.X+wheel.bounds.Width)
	wheel.selectPos[1] = math2.Clamp(wheel.selectPos[1], wheel.bounds.Y, wheel.bounds.Y+wheel.bounds.Height)

	wheel.cursor.SetDestPosition(wheel.selectPos)
	wheel.cursor.Color.A = openness
	queue.Add(&wheel.cursor)

	slotTexture := cache.GetTexture(texWeaponSlot)
	slotWidth := slotTexture.Rect().Width * settings.SpriteScale()
	slotHeight := slotTexture.Rect().Height * settings.SpriteScale()
	majorRadiusSq := math2.Pow(slotWidth/2.0, 2.0)
	minorRadiusSq := math2.Pow(slotHeight/2.0, 2.0)

	for x := range wheel.slots {
		for y := range wheel.slots[x] {
			slot := &wheel.slots[x][y]

			if slot.kind != WeaponSickle {
				// Move slot towards target position
				dx := (slot.targetPos[0] - slot.back.Dest.X) * 0.5
				dy := (slot.targetPos[1] - slot.back.Dest.Y) * 0.5
				slot.back.Dest.X += dx
				slot.back.Dest.Y += dy
				slot.icon.Dest.X += dx
				slot.icon.Dest.Y += dy
			}

			// Test intersection with the ellipse
			cx, cy := slot.back.Dest.Center()
			if (math2.Pow(wheel.selectPos[0]-cx, 2.0)/majorRadiusSq)+(math2.Pow(wheel.selectPos[1]-cy, 2.0)/minorRadiusSq) <= 1.0 && wheel.highlightedWeapon != slot.kind {
				wheel.highlightedWeapon = slot.kind
				cache.GetSfx("assets/sounds/ui/weapon_select.wav").Play()
			}

			if wheel.highlightedWeapon == slot.kind {
				wheel.highlight.SetDestPosition(slot.back.DestPosition())
				wheel.highlight.Color.A = openness
				queue.Add(&wheel.highlight)
			}

			slot.back.Color.A = openness

			queue.Add(&slot.back)

			if openness > 0.9 {
				queue.Add(&slot.icon)
			}
		}
	}
}
