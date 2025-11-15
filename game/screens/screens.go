package screens

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

type Screen struct {
	position      mgl32.Vec2
	menuSelection int
	menuTexts     []ui.Text
	menuSpacing   float32
	cursor        ui.Box
	onSelect      func(item string)
}

func (scr *Screen) init(position mgl32.Vec2, menuItems []string) {
	*scr = Screen{
		position:  position,
		menuTexts: make([]ui.Text, len(menuItems)),
	}

	scr.menuSpacing = 24 * settings.UIScale()

	scr.cursor = ui.NewBoxFull(math2.Rect{
		Width: 16, Height: 16,
	}, cache.GetTexture("assets/textures/ui/menu_cursor.png"), color.White, 10.0)

	for i, text := range menuItems {
		scr.menuTexts[i] = ui.Text{
			Transform: ui.Transform{
				Dest: math2.Rect{
					X:      position[0] + (24.0 * settings.UIScale()),
					Y:      position[1] + (scr.menuSpacing * float32(i)),
					Width:  (settings.UIWidth() - position[0] - 24.0),
					Height: scr.menuSpacing,
				},
			},
			Settings: ui.TextSettings{
				Text:         text,
				ShadowColor:  color.Black,
				ShadowOffset: mgl32.Vec2{4.0, 4.0},
			},
		}
		scr.menuTexts[i].SetText(text)
	}
}

func (scr *Screen) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if input.IsActionJustPressed(settings.ActionMenuDown) {
		scr.menuSelection = (scr.menuSelection + 1) % len(scr.menuTexts)
	} else if input.IsActionJustPressed(settings.ActionMenuUp) {
		scr.menuSelection = (scr.menuSelection + len(scr.menuTexts) - 1) % len(scr.menuTexts)
	}

	if scr.onSelect != nil && input.IsActionJustPressed(settings.ActionMenuConfirm) {
		scr.onSelect(scr.menuTexts[scr.menuSelection].Text())
	}

	for i := range scr.menuTexts {
		elem := &scr.menuTexts[i]
		if scr.menuSelection == i {
			elem.Color = color.Yellow
		} else {
			elem.Color = color.White
		}
		queue.Add(elem)
	}

	// Draw cursor in front of menu items
	scr.cursor.SetDestPosition(mgl32.Vec2{
		scr.position[0],
		scr.position[1] + (scr.menuSpacing * float32(scr.menuSelection)) + (5.0 * settings.UIScale()),
	})
	scr.cursor.Transform.Rotation -= deltaTime * math.Pi * 4.0
	queue.Add(&scr.cursor)
}
