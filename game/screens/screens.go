package screens

import (
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Screen struct {
	position      mgl32.Vec2
	menuSelection int
	menuTexts     []ui.Element
	menuSpacing   float32
	cursor        ui.Element
	onSelect      func(item string) bool
	app           engine.Observer
}

func (scr *Screen) init(app engine.Observer, position mgl32.Vec2, menuItems []string) {
	*scr = Screen{
		position:      position,
		menuTexts:     make([]ui.Element, len(menuItems)),
		menuSelection: -1,
		app:           app,
	}

	scr.menuSpacing = 24 * settings.UIScale()

	scr.cursor = ui.NewBox(ui.Transform{
		Size:   mgl32.Vec2{16.0, 16.0},
		Origin: ui.Ratios{0.5, 0.5},
		Depth:  10,
	}, cache.GetTexture("assets/textures/ui/menu_cursor.png"))

	for i, text := range menuItems {
		scr.menuTexts[i] = ui.Element{
			Transform: ui.Transform{
				Position: mgl32.Vec2{
					position[0] + (24.0 * settings.UIScale()),
					position[1] + (scr.menuSpacing * float32(i)),
				},
				Size: mgl32.Vec2{
					(settings.UIWidth() - position[0] - 24.0),
					24.0,
				},
				Origin: ui.Ratios{0.0, 0.5},
			},
			ShadowColor:  color.Black,
			ShadowOffset: mgl32.Vec2{4.0, 4.0},
		}
		scr.menuTexts[i].SetText(text)
		scr.menuTexts[i].ShrinkToFitText()
	}
}

func (scr *Screen) Layout(queue *ui.RenderQueue, deltaTime float32) bool {
	if input.IsActionJustPressed(settings.ActionMenuDown) {
		scr.menuSelection = (scr.menuSelection + 1) % len(scr.menuTexts)
	} else if input.IsActionJustPressed(settings.ActionMenuUp) {
		scr.menuSelection = (scr.menuSelection + len(scr.menuTexts) - 1) % len(scr.menuTexts)
	}

	if scr.onSelect != nil && scr.menuSelection >= 0 && input.IsActionJustPressed(settings.ActionMenuConfirm) {
		return scr.onSelect(scr.menuTexts[scr.menuSelection].Text())
	}

	if input.MouseDelta().LenSqr() > 0.1 {
		scr.menuSelection = -1
	}

	for i := range scr.menuTexts {
		elem := &scr.menuTexts[i]
		if elem.OnScreenBox().ContainsPoint(input.MousePosition()) {
			scr.menuSelection = i
			if scr.onSelect != nil && (input.IsMouseButtonDown(glfw.MouseButton1) || input.IsMouseButtonDown(glfw.MouseButton2)) {
				return scr.onSelect(elem.Text())
			}
		}
		if scr.menuSelection == i {
			elem.Color = color.Yellow
		} else {
			elem.Color = color.White
		}
		queue.Add(elem)
	}

	if scr.menuSelection >= 0 {
		// Draw cursor in front of menu items
		scr.cursor.Position = mgl32.Vec2{
			scr.position[0],
			scr.position[1] + (scr.menuSpacing * float32(scr.menuSelection)),
		}
		scr.cursor.Rotation -= math2.Radians(deltaTime * math.Pi * 4.0)
		queue.Add(&scr.cursor)
	}

	return true
}
