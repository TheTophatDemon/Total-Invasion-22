package screens

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Menu struct {
	position      mgl32.Vec2
	menuSelection int
	menuTexts     []ui.Element
	menuSpacing   float32
	cursor        ui.Element
	title, fade   ui.Element
	onSelect      func(itemIndex int)
	app           engine.Observer
}

func newMenu(app engine.Observer, position mgl32.Vec2, menuItems []string) *Menu {
	menu := Menu{
		position:      position,
		menuTexts:     make([]ui.Element, len(menuItems)),
		menuSelection: -1,
		app:           app,
	}

	menu.menuSpacing = 36
	menu.cursor = ui.NewBox(ui.Transform{
		Size:   mgl32.Vec2{32.0, 32.0},
		Origin: ui.Ratios{0.5, 0.5},
		Depth:  10,
	}, cache.GetTexture("assets/textures/ui/menu_cursor.png"))

	for i, text := range menuItems {
		menu.menuTexts[i] = ui.NewText(
			ui.Transform{
				Position: mgl32.Vec2{
					position[0] + (36),
					position[1] + (menu.menuSpacing * float32(i)),
				},
				Size: mgl32.Vec2{
					(settings.UIWidth() - position[0] - 24.0),
					24.0,
				},
				Origin: ui.Ratios{0.0, 0.5},
				Depth:  10,
			},
			settings.Localize(text),
			color.White,
		)
		menu.menuTexts[i].ShrinkToFitText()
	}

	menu.title = ui.NewBox(ui.Transform{
		Size:   mgl32.Vec2{960.0, 256.0},
		Anchor: ui.Ratios{0.5, 0.0},
		Origin: ui.Ratios{0.5, 0.0},
		Depth:  50,
	}, cache.GetTexture("assets/textures/ui/title.png"))

	menu.fade = ui.NewBox(ui.Transform{
		Size:  mgl32.Vec2{float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight)},
		Depth: 1,
	}, nil)
	menu.fade.Color = color.Black.WithAlpha(0.25)

	return &menu
}

func (menu *Menu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	queue.Add(&menu.fade)

	if input.IsActionJustPressed(settings.ActionMenuDown) {
		menu.menuSelection = (menu.menuSelection + 1) % len(menu.menuTexts)
	} else if input.IsActionJustPressed(settings.ActionMenuUp) {
		menu.menuSelection = (menu.menuSelection + len(menu.menuTexts) - 1) % len(menu.menuTexts)
	}

	if menu.onSelect != nil && menu.menuSelection >= 0 && input.IsActionJustPressed(settings.ActionMenuConfirm) {
		menu.onSelect(menu.menuSelection)
	}

	if input.MouseDelta().LenSqr() > 0.1 {
		menu.menuSelection = -1
	}

	for i := range menu.menuTexts {
		elem := &menu.menuTexts[i]
		if elem.OnScreenBox().ContainsPoint(input.MousePosition()) {
			menu.menuSelection = i
			if menu.onSelect != nil && input.IsActionJustReleased(settings.ActionMenuClick) {
				menu.onSelect(i)
			}
		}
		if menu.menuSelection == i {
			elem.Color = color.Yellow
		} else {
			elem.Color = color.White
		}
		queue.Add(elem)
	}

	if menu.menuSelection >= 0 {
		// Draw cursor in front of menu items
		menu.cursor.Position = mgl32.Vec2{
			menu.position[0],
			menu.position[1] + (menu.menuSpacing * float32(menu.menuSelection)),
		}
		menu.cursor.Rotation -= math2.Radians(deltaTime * math.Pi * 4.0)
		queue.Add(&menu.cursor)
	}

	queue.Add(&menu.title)
}
