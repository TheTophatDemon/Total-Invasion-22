package screens

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
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
	MenuInputType = uint8
	MenuWidget    interface {
		Element() *ui.Element
		Focus()
		Blur()
		Input(MenuInputType, *Menu)
		Layout(queue *ui.RenderQueue, deltaTime float32)
	}
	Menu struct {
		position                                 mgl32.Vec2
		scrollY, targetScrollY                   float32
		menuSelection                            int
		menuItems                                []MenuWidget
		menuSpacing                              float32
		cursor, scrollUpButton, scrollDownButton ui.Element
		title, fade                              ui.Element
		blinker                                  gween.Sequence
		app                                      engine.Observer
		parent                                   ui.Screen
	}
	element  = ui.Element // This allows us to embed the type privately without colliding with the Element() method
	MenuItem struct {
		element
		OnInput      func(MenuInputType)
		restoreColor maybe.T[color.Color]
	}
	ReturnItem struct {
		MenuItem
	}
)

const (
	MenuInputConfirm MenuInputType = iota
	MenuInputIncrement
	MenuInputDecrement
	MenuInputClick
)

const SfxMenuMove = "assets/sounds/ui/menu0.wav"
const SfxMenuHit = "assets/sounds/ui/menu1.wav"

func (menu *Menu) Init(app engine.Observer, menuItems []MenuWidget, parent ui.Screen) *Menu {
	*menu = Menu{
		menuSpacing:   36,
		menuItems:     menuItems,
		menuSelection: 0,
		app:           app,
		parent:        parent,
		blinker: *gween.NewSequence( // Blink on and off
			gween.New(1.0, 1.0, 0.75, ease.Linear),
			gween.New(0.0, 0.0, 0.25, ease.Linear),
		),
	}

	smallScreen := settings.Current.WindowWidth <= 800

	menu.blinker.SetLoop(-1)

	menu.position = mgl32.Vec2{
		float32(settings.Current.WindowWidth) * 0.1,
		float32(settings.Current.WindowHeight) * 0.4,
	}

	if smallScreen {
		menu.menuSpacing = 48
		menu.position[0] = 16
	}

	menu.cursor = ui.NewBox(ui.Transform{
		Size:     mgl32.Vec2{32.0, 32.0},
		Origin:   ui.Ratios{0.5, 0.5},
		Depth:    10,
		Position: mgl32.Vec2{-2050.0, -2050.0}, // Spawn offscreen until menu item is selected
	}, cache.GetTexture("assets/textures/ui/menu_cursor.png"))

	menu.scrollUpButton = ui.NewBox(ui.Transform{
		Size:     mgl32.Vec2{64.0, 16.0},
		Origin:   ui.Ratios{0.5, 1.0},
		Depth:    10,
		Anchor:   ui.Ratios{0.5, 0.0},
		Position: mgl32.Vec2{0.0, menu.position[1]},
	}, cache.GetTexture("assets/textures/ui/scroll_arrow.png"))

	menu.scrollDownButton = ui.NewBox(ui.Transform{
		Size:     mgl32.Vec2{64.0, 16.0},
		Origin:   ui.Ratios{0.5, 0.0},
		Depth:    10,
		Anchor:   ui.Ratios{0.5, 0.0},
		Position: mgl32.Vec2{0.0, float32(settings.Current.WindowHeight) - menu.menuSpacing},
	}, cache.GetTexture("assets/textures/ui/scroll_arrow_down.png"))

	for i := range menuItems {
		elem := menu.menuItems[i].Element()
		if elem == nil {
			continue
		}
		elem.SetTransform(
			ui.Transform{
				Position: mgl32.Vec2{
					menu.position[0] + 36,
					menu.position[1] + (menu.menuSpacing * float32(i+1)),
				},
				Size: mgl32.Vec2{
					(settings.UIWidth() - menu.position[0] - 24.0),
					24.0,
				},
				Origin: ui.Ratios{0.0, 0.5},
				Depth:  10,
			})
		elem.FitText()
		if smallScreen && elem.Size()[0] >= 300 {
			elem.SetSize(mgl32.Vec2{300, 48})
		}

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

func (menu *Menu) scrollUp() {
	cache.GetSfx(sfxChoose).Play()
	menu.targetScrollY = min(0.0, menu.targetScrollY+menu.menuSpacing)
}

func (menu *Menu) scrollDown(minScrollY float32) {
	cache.GetSfx(sfxChoose).Play()
	menu.targetScrollY = max(minScrollY, menu.targetScrollY-menu.menuSpacing)
}

func (menu *Menu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	queue.Add(&menu.fade)

	prevSelection := menu.menuSelection
	if settings.Current.ActionMenuDown.JustPressed() {
		menu.menuSelection = (menu.menuSelection + 1) % len(menu.menuItems)
	} else if settings.Current.ActionMenuUp.JustPressed() {
		menu.menuSelection = (menu.menuSelection + len(menu.menuItems) - 1) % len(menu.menuItems)
	}

	if settings.Current.ActionMenuCancel.JustPressed() {
		menu.returnToParent()
	}

	if menu.menuSelection >= 0 {
		item := menu.menuItems[menu.menuSelection]
		if settings.Current.ActionMenuConfirm.JustPressed() {
			item.Input(MenuInputConfirm, menu)
		} else if settings.Current.ActionMenuIncrement.JustPressed() {
			item.Input(MenuInputIncrement, menu)
		} else if settings.Current.ActionMenuDecrement.JustPressed() {
			item.Input(MenuInputDecrement, menu)
		}
	}

	scroll := input.MouseScrollDelta()
	mouseMoved := input.MouseDelta().LenSqr() > 0.1 || !mgl32.FloatEqual(scroll, 0.0)
	if mouseMoved {
		menu.menuSelection = -1
	}

	menuTopY := menu.position[1] + menu.menuSpacing
	menuBottomY := float32(settings.Current.WindowHeight) - menu.menuSpacing
	moreAbove := false
	moreBelow := false

	menuItemsHeight := float32(0)
	if numItems := len(menu.menuItems); numItems > 0 {
		menuItemsHeight = (menu.menuSpacing * float32(numItems)) + menu.menuItems[numItems-1].Element().Size()[1]
	}
	minScrollY := menuBottomY - menuTopY - menuItemsHeight

	// Draw menu items and take user input
	for i := range menu.menuItems {
		item := menu.menuItems[i]
		elem := item.Element()

		elem.SetPosition(mgl32.Vec2{
			menu.position[0] + 36,
			menu.position[1] + (menu.menuSpacing * float32(i+1)) + menu.scrollY,
		})

		noDraw := false
		// Occlude menu items outside of the viewable range
		if elem.Position()[1]+elem.Size()[1] < menuTopY {
			moreAbove = true
			noDraw = true
		}
		if elem.Position()[1]+elem.Size()[1] > menuBottomY {
			moreBelow = true
			noDraw = true
		}
		if !noDraw && elem.OnScreenBox().ContainsPoint(input.MousePosition()) {
			if mouseMoved {
				menu.menuSelection = i
			}
			if menu.menuSelection == i && input.AnyMouseButtonJustPressed() {
				item.Input(MenuInputClick, menu)
			}
		}
		if menu.menuSelection == i && prevSelection != i {
			item.Focus()
			if noDraw {
				// Scroll off screen menu item into view
				menu.targetScrollY = min(0.0, max(minScrollY, -menu.menuSpacing*float32(i)))
			}
		} else if menu.menuSelection != i && prevSelection == i {
			item.Blur()
		}
		if !noDraw {
			item.Layout(queue, deltaTime)
		}
	}

	// Draw scroller arrows
	scrollerValue, _, _ := menu.blinker.Update(deltaTime)
	showScrollers := scrollerValue > 0.9
	if moreAbove {
		menu.scrollUpButton.SetPosition(mgl32.Vec2{0.0, menu.position[1]})
		if menu.scrollUpButton.OnScreenBox().ContainsPoint(input.MousePosition()) {
			menu.scrollUpButton.BgColor = maybe.Some(color.Yellow)
			if input.AnyMouseButtonJustPressed() {
				menu.scrollUp()
			}
		} else {
			menu.scrollUpButton.BgColor = maybe.Some(color.White)
		}
		if showScrollers {
			queue.Add(&menu.scrollUpButton)
		}
	}

	if moreBelow {
		menu.scrollDownButton.SetPosition(mgl32.Vec2{0.0, menuBottomY})
		if menu.scrollDownButton.OnScreenBox().ContainsPoint(input.MousePosition()) {
			menu.scrollDownButton.BgColor = maybe.Some(color.Yellow)
			if input.AnyMouseButtonJustPressed() {
				menu.scrollDown(minScrollY)
			}
		} else {
			menu.scrollDownButton.BgColor = maybe.Some(color.White)
		}
		if showScrollers {
			queue.Add(&menu.scrollDownButton)
		}
	}

	if (moreAbove || moreBelow) && !mgl32.FloatEqual(scroll, 0.0) {
		for ; scroll > 0.0; scroll -= 1.0 {
			menu.scrollUp()
		}
		for ; scroll < 0.0; scroll += 1.0 {
			menu.scrollDown(minScrollY)
		}
		// The cursor will end up in a weird position, so just hide it for now
		menu.cursor.SetPosition(mgl32.Vec2{-2048.0, -2048.0})
	}

	// Move scroll value
	if !mgl32.FloatEqual(menu.scrollY, menu.targetScrollY) {
		menu.scrollY += (menu.targetScrollY - menu.scrollY) * 0.5
	}

	if menu.menuSelection >= 0 {
		// Draw cursor in front of menu items
		target := mgl32.Vec2{
			menu.position[0],
			menu.menuItems[menu.menuSelection].Element().Position()[1],
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

func (item *MenuItem) Init(stringKey string, callback func(MenuInputType)) *MenuItem {
	if item == nil {
		return nil
	}
	return item.InitUnlocalized(settings.Localize(stringKey), callback)
}

func (item *MenuItem) InitUnlocalized(actualText string, callback func(MenuInputType)) *MenuItem {
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
	}
}

func (item *MenuItem) Blur() {
	configMaybe := item.element.TextConfig()
	if config, ok := configMaybe.Get(); ok {
		config.Color = item.restoreColor
	}
	cache.GetSfx(SfxMenuMove).Play()
}

func (item *MenuItem) Input(action MenuInputType, menu *Menu) {
	if (action == MenuInputConfirm || action == MenuInputClick) && item.OnInput != nil {
		item.OnInput(action)
		cache.GetSfx(SfxMenuHit).Play()
	}
}

func (item *MenuItem) Layout(queue *ui.RenderQueue, deltaTime float32) {
	queue.Add(&item.element)
}

func (returnItem *ReturnItem) Init() *ReturnItem {
	if returnItem == nil {
		return nil
	}
	*returnItem = ReturnItem{
		MenuItem: MenuItem{
			element: ui.NewText(ui.Transform{}, settings.Localize("return"), ui.TextConfig{}),
		},
	}
	return returnItem
}

func (returnItem *ReturnItem) Input(action MenuInputType, menu *Menu) {
	if menu == nil {
		return
	}
	if action == MenuInputConfirm || action == MenuInputClick {
		cache.GetSfx(SfxMenuHit).Play()
		menu.returnToParent()
	}
}
