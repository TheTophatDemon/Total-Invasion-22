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

type (
	MenuItem struct {
		Element  ui.Element
		OnSelect func(menu *Menu)
	}
	Menu struct {
		position      mgl32.Vec2
		menuSelection int
		menuItems     []MenuItem
		menuSpacing   float32
		cursor        ui.Element
		title, fade   ui.Element
		app           engine.Observer
	}
)

func newMenu(app engine.Observer, position mgl32.Vec2, menuItems []MenuItem) *Menu {
	menu := Menu{
		position:      position,
		menuItems:     menuItems,
		menuSelection: -1,
		app:           app,
	}

	menu.menuSpacing = 36
	menu.cursor = ui.NewBox(ui.Transform{
		Size:   mgl32.Vec2{32.0, 32.0},
		Origin: ui.Ratios{0.5, 0.5},
		Depth:  10,
	}, cache.GetTexture("assets/textures/ui/menu_cursor.png"))

	for i := range menuItems {
		menu.menuItems[i].Element.SetTransform(
			ui.Transform{
				Position: mgl32.Vec2{
					position[0] + 36,
					position[1] + (menu.menuSpacing * float32(i)),
				},
				Size: mgl32.Vec2{
					(settings.UIWidth() - position[0] - 24.0),
					24.0,
				},
				Origin: ui.Ratios{0.0, 0.5},
				Depth:  10,
			})
		menu.menuItems[i].Element.ShrinkToFitText()
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
		menu.menuSelection = (menu.menuSelection + 1) % len(menu.menuItems)
	} else if input.IsActionJustPressed(settings.ActionMenuUp) {
		menu.menuSelection = (menu.menuSelection + len(menu.menuItems) - 1) % len(menu.menuItems)
	}

	if menu.menuSelection >= 0 && input.IsActionJustPressed(settings.ActionMenuConfirm) {
		item := &menu.menuItems[menu.menuSelection]
		if item.OnSelect != nil {
			item.OnSelect(menu)
		}
	}

	if input.MouseDelta().LenSqr() > 0.1 {
		menu.menuSelection = -1
	}

	for i := range menu.menuItems {
		item := &menu.menuItems[i]
		if item.Element.OnScreenBox().ContainsPoint(input.MousePosition()) {
			menu.menuSelection = i
			if item.OnSelect != nil && input.IsActionJustReleased(settings.ActionMenuClick) {
				item.OnSelect(menu)
			}
		}
		if menu.menuSelection == i {
			item.Element.Color = color.Yellow
		} else {
			item.Element.Color = color.White
		}
		queue.Add(&item.Element)
	}

	if menu.menuSelection >= 0 {
		// Draw cursor in front of menu items
		menu.cursor.SetPosition(mgl32.Vec2{
			menu.position[0],
			menu.position[1] + (menu.menuSpacing * float32(menu.menuSelection)),
		})
		menu.cursor.Rotate(-math2.Radians(deltaTime * math.Pi * 4.0))
		queue.Add(&menu.cursor)
	}

	queue.Add(&menu.title)
}
