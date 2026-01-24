package screens

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	MenuEvents interface {
		Element() *ui.Element
		Focus()
		Blur()
		Input(action input.Action)
		Layout(queue *ui.RenderQueue, deltaTime float32)
	}
	Menu struct {
		position      mgl32.Vec2
		menuSelection int
		menuItems     []MenuEvents
		menuSpacing   float32
		cursor        ui.Element
		title, fade   ui.Element
		app           engine.Observer
		parent        ui.Screen
	}
	element  = ui.Element // This allows us to embed the type privately without colliding with the Element() method
	MenuItem struct {
		element
		OnInput      func(action input.Action)
		restoreColor maybe.T[color.Color]
	}
)

func (menu *Menu) Init(app engine.Observer, menuItems []MenuEvents, parent ui.Screen) *Menu {
	*menu = Menu{
		menuSpacing:   36,
		menuItems:     menuItems,
		menuSelection: 0,
		app:           app,
		parent:        parent,
	}

	menu.position = mgl32.Vec2{
		(float32(settings.Current.WindowWidth) / 2.0) - 256.0,
		float32(int(settings.Current.WindowHeight) - ((len(menuItems) + 1) * int(menu.menuSpacing))),
	}

	menu.cursor = ui.NewBox(ui.Transform{
		Size:     mgl32.Vec2{32.0, 32.0},
		Origin:   ui.Ratios{0.5, 0.5},
		Depth:    10,
		Position: mgl32.Vec2{-2050.0, -2050.0}, // Spawn offscreen until menu item is selected
	}, cache.GetTexture("assets/textures/ui/menu_cursor.png"))

	for i := range menuItems {
		elem := menu.menuItems[i].Element()
		if elem == nil {
			continue
		}
		elem.SetTransform(
			ui.Transform{
				Position: mgl32.Vec2{
					menu.position[0] + 36,
					menu.position[1] + (menu.menuSpacing * float32(i)),
				},
				Size: mgl32.Vec2{
					(settings.UIWidth() - menu.position[0] - 24.0),
					24.0,
				},
				Origin: ui.Ratios{0.0, 0.5},
				Depth:  10,
			})
		elem.FitText()
		if i == 0 {
			menu.menuItems[0].Focus()
		}
	}

	menu.title = ui.NewBox(ui.Transform{
		Anchor:   ui.Ratios{0.5, 0.0},
		Origin:   ui.Ratios{0.5, 0.0},
		Position: mgl32.Vec2{0.0, 0.0},
		Depth:    50,
	}, cache.GetTexture("assets/textures/ui/title.png"))
	menu.title.FitHeight(float32(settings.Current.WindowHeight) * 0.355555556)

	menu.fade = ui.NewBox(ui.Transform{
		Size:  mgl32.Vec2{float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight)},
		Depth: 1,
	}, nil)
	menu.fade.BgColor = maybe.Some(color.Black.WithAlpha(0.25))

	return menu
}

func (menu *Menu) returnToParent() {
	menu.app.ProcessSignal(game.ChangeScreenSignal{
		Screen: menu.parent,
	})
}

func (menu *Menu) Enter() {
}

func (menu *Menu) Exit() {
}

func (menu *Menu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	queue.Add(&menu.fade)

	prevSelection := menu.menuSelection
	if input.IsActionJustPressed(settings.ActionMenuDown) {
		menu.menuSelection = (menu.menuSelection + 1) % len(menu.menuItems)
	} else if input.IsActionJustPressed(settings.ActionMenuUp) {
		menu.menuSelection = (menu.menuSelection + len(menu.menuItems) - 1) % len(menu.menuItems)
	}

	if input.IsActionJustPressed(settings.ActionMenuCancel) {
		menu.returnToParent()
	}

	if menu.menuSelection >= 0 {
		item := menu.menuItems[menu.menuSelection]
		if input.IsActionJustPressed(settings.ActionMenuConfirm) {
			item.Input(settings.ActionMenuConfirm)
		} else if input.IsActionJustPressed(settings.ActionMenuIncrement) {
			item.Input(settings.ActionMenuIncrement)
		} else if input.IsActionJustPressed(settings.ActionMenuDecrement) {
			item.Input(settings.ActionMenuDecrement)
		}
	}

	mouseMoved := input.MouseDelta().LenSqr() > 0.1
	if mouseMoved {
		menu.menuSelection = -1
	}

	for i := range menu.menuItems {
		item := menu.menuItems[i]
		if elem := item.Element(); elem != nil && elem.OnScreenBox().ContainsPoint(input.MousePosition()) {
			if mouseMoved {
				menu.menuSelection = i
			}
			if menu.menuSelection == i && input.IsActionJustReleased(settings.ActionMenuClick) {
				item.Input(settings.ActionMenuClick)
			}
		}
		if menu.menuSelection == i && prevSelection != i {
			item.Focus()
		} else if menu.menuSelection != i && prevSelection == i {
			item.Blur()
		}
		item.Layout(queue, deltaTime)
	}

	if menu.menuSelection >= 0 {
		// Draw cursor in front of menu items
		target := mgl32.Vec2{
			menu.position[0],
			menu.position[1] + (menu.menuSpacing * float32(menu.menuSelection)),
		}
		diff := target.Sub(menu.cursor.Position())
		if diff.Len() > 2048.0 {
			menu.cursor.SetPosition(target)
		} else {
			menu.cursor.Translate(diff.Mul(0.5))
		}
	}

	menu.cursor.Rotate(-math2.Radians(deltaTime * math.Pi * 4.0))
	queue.Add(&menu.cursor)
	queue.Add(&menu.title)
}

func (item *MenuItem) Init(stringKey string, callback func(input.Action)) *MenuItem {
	if item == nil {
		return nil
	}
	return item.InitUnlocalized(settings.Localize(stringKey), callback)
}

func (item *MenuItem) InitUnlocalized(actualText string, callback func(input.Action)) *MenuItem {
	if item == nil {
		return nil
	}
	item.element = ui.NewText(ui.Transform{}, actualText, ui.TextConfig{})
	item.OnInput = callback
	return item
}

func (item *MenuItem) Element() *ui.Element {
	return &item.element
}

func (item *MenuItem) Focus() {
	configMaybe := item.element.TextConfig()
	if config, ok := configMaybe.Get(); ok {
		item.restoreColor = config.Color
		config.Color = maybe.Some(color.Yellow)
		item.element.SetTextConfig(*config)
	}
}

func (item *MenuItem) Blur() {
	configMaybe := item.element.TextConfig()
	if config, ok := configMaybe.Get(); ok {
		config.Color = item.restoreColor
		item.element.SetTextConfig(*config)
	}
	cache.GetSfx("assets/sounds/ui/menu0.wav").Play()
}

func (item *MenuItem) Input(action input.Action) {
	if (action == settings.ActionMenuConfirm || action == settings.ActionMenuClick) && item.OnInput != nil {
		item.OnInput(action)
		cache.GetSfx("assets/sounds/ui/menu1.wav").Play()
	}
}

func (item *MenuItem) Layout(queue *ui.RenderQueue, deltaTime float32) {
	queue.Add(&item.element)
}
