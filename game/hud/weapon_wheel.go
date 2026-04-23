package hud

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const texWeaponSlot = "assets/textures/ui/weapon_slot.png"
const wheelIconTransparency = 0.2

type weaponSlot struct {
	kind       game.WeaponType
	back, icon ui.Element
	targetPos  mgl32.Vec2 // The position where each slot will move towards
}

type WeaponWheel struct {
	slots             [3][3]weaponSlot
	highlight, cursor ui.Element
	highlightedWeapon game.WeaponType
	selectPos         mgl32.Vec2 // Represents a virtual mouse position for selecting the weapon
	bounds            math2.Rect // Rectangular region within which the wheel resides
}

func newWeaponWheel(weapons []Weapon) WeaponWheel {
	const slotMargin = 16.0
	slotTexture := cache.GetTexture(texWeaponSlot)
	slotWidth := slotTexture.Rect().Width * SpriteScale()
	slotHeight := slotTexture.Rect().Height * SpriteScale()
	slotSize := mgl32.Vec2{slotWidth, slotHeight}

	wheel := WeaponWheel{
		selectPos: mgl32.Vec2{settings.UIWidth() / 2.0, settings.UIHeight() / 2.0},
		bounds: math2.Rect{
			X:      (settings.UIWidth() / 2.0) - slotMargin - (slotWidth * 1.5),
			Y:      (settings.UIHeight() / 2.0) - slotMargin - (slotHeight * 1.5),
			Width:  (slotWidth * 3.0) + (slotMargin * 2.0),
			Height: (slotHeight * 3.0) + (slotMargin * 2.0),
		},
		highlightedWeapon: game.WeaponSickle,
	}

	slotStart := wheel.selectPos.Sub(mgl32.Vec2{slotWidth / 2.0, slotHeight / 2.0})

	for i, kind := range [...]game.WeaponType{
		game.WeaponCluckster, game.WeaponChicken, game.WeaponAirhorn,
		game.WeaponSign, game.WeaponSickle, game.WeaponGrenade,
		game.WeaponDefenestrator, game.WeaponParusu, game.WeaponDblGrenade,
	} {
		weapon := weapons[kind]

		depth := float32(6.0)
		if kind == game.WeaponSickle {
			depth = 6.1 // Put sickle above other slots while they are expanding out
		}

		var icon ui.Element
		if iconPath := weapon.wheelIconPath; len(iconPath) > 0 {
			iconTex := cache.GetTexture(iconPath)
			iconWidth := min(iconTex.Rect().Width*SpriteScale(), 40*SpriteScale())
			iconHeight := min(iconTex.Rect().Height*SpriteScale(), 40*SpriteScale())

			icon = ui.NewBox(ui.Transform{
				Position: mgl32.Vec2{
					slotStart[0] + (slotWidth-iconWidth)/2.0,
					slotStart[1] + (slotHeight-iconHeight)/2.0,
				},
				Size:  mgl32.Vec2{iconWidth, iconHeight},
				Depth: 7.5,
			}, iconTex)

			if !weapon.Equipped {
				// Intentionally overflow the color values to erase detail on the image.
				icon.BgColor = maybe.Some(color.Color{R: 100.0, G: 100.0, B: 100.0, A: wheelIconTransparency})
			}
		}

		x, y := i%len(wheel.slots[0]), i/len(wheel.slots[0])

		slot := weaponSlot{
			kind: kind,
			back: ui.NewBox(ui.Transform{
				// All slots start in the center and move towards their target positions
				Position: slotStart,
				Size:     slotSize,
				Depth:    depth,
			}, slotTexture),
			icon: icon,
			targetPos: mgl32.Vec2{
				wheel.bounds.X + float32(x)*(slotWidth+slotMargin),
				wheel.bounds.Y + float32(y)*(slotHeight+slotMargin),
			},
		}
		slot.back.BgColor = maybe.Some(weapon.wheelColor)

		wheel.slots[x][y] = slot
	}

	cursorTex := cache.GetTexture("assets/textures/ui/hand_cursor.png")
	wheel.cursor = ui.NewBox(
		ui.Transform{
			Size: mgl32.Vec2{
				cursorTex.Rect().Width * SpriteScale(),
				cursorTex.Rect().Height * SpriteScale(),
			},
			Depth: 8.0,
		},
		cursorTex,
	)

	wheel.highlight = ui.NewBox(ui.Transform{
		Position: mgl32.Vec2{
			(settings.UIWidth() / 2.0) - (slotWidth / 2.0),
			(settings.UIHeight() / 2.0) - (slotHeight / 2.0),
		},
		Size:  slotSize,
		Depth: 7.0,
	}, slotTexture)

	return wheel
}

func (wheel *WeaponWheel) Layout(queue *ui.RenderQueue, openness float32) {
	wheel.selectPos = wheel.selectPos.Add(input.MouseDelta())

	wheel.selectPos[0] = math2.Clamp(wheel.selectPos[0], wheel.bounds.X, wheel.bounds.X+wheel.bounds.Width)
	wheel.selectPos[1] = math2.Clamp(wheel.selectPos[1], wheel.bounds.Y, wheel.bounds.Y+wheel.bounds.Height)

	wheel.cursor.SetPosition(wheel.selectPos)
	wheel.cursor.BgColor = maybe.Some(color.White.WithAlpha(openness))
	queue.Add(&wheel.cursor)

	slotTexture := cache.GetTexture(texWeaponSlot)
	slotWidth := slotTexture.Rect().Width * SpriteScale()
	slotHeight := slotTexture.Rect().Height * SpriteScale()
	majorRadiusSq := math2.Pow(slotWidth/2.0, 2.0)
	minorRadiusSq := math2.Pow(slotHeight/2.0, 2.0)

	for x := range wheel.slots {
		for y := range wheel.slots[x] {
			slot := &wheel.slots[x][y]

			if slot.kind != game.WeaponSickle {
				// Move slot towards target position
				delta := mgl32.Vec2{
					(slot.targetPos[0] - slot.back.X()) * 0.5,
					(slot.targetPos[1] - slot.back.Y()) * 0.5,
				}
				slot.back.Translate(delta)
				slot.icon.Translate(delta)
			}

			// Test intersection with the ellipse
			center := slot.back.Center()
			intersects := (math2.Pow(wheel.selectPos[0]-center[0], 2.0)/majorRadiusSq)+(math2.Pow(wheel.selectPos[1]-center[1], 2.0)/minorRadiusSq) <= 1.0
			if intersects && wheel.highlightedWeapon != slot.kind {
				wheel.highlightedWeapon = slot.kind
				cache.GetSfx("assets/sounds/ui/weapon_select.wav").Play()
			}

			if wheel.highlightedWeapon == slot.kind {
				wheel.highlight.SetPosition(slot.back.Position())
				if clr, ok := wheel.highlight.BgColor.Get(); ok {
					clr.A = openness
				}
				queue.Add(&wheel.highlight)
			}

			if clr, ok := slot.back.BgColor.Get(); ok {
				clr.A = openness
			}

			queue.Add(&slot.back)

			if openness > 0.9 {
				queue.Add(&slot.icon)
			}
		}
	}
}
