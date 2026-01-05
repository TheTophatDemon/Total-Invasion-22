package screens

import (
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

func NewTitleMenu(app engine.Observer, inGame bool) *Menu {
	pos := mgl32.Vec2{
		(float32(settings.Current.WindowWidth) / 2.0) - 256.0,
		float32(settings.Current.WindowHeight) - 192.0,
	}
	var titleScreenMenus = [...]MenuItem{
		{
			Element: ui.NewText(ui.Transform{}, settings.Localize("resumeGame"), ui.TextConfig{}),
			OnSelect: func(menu *Menu) {
				menu.app.ProcessSignal(game.ResumeGameSignal{})
			},
		},
		{
			Element: ui.NewText(ui.Transform{}, settings.Localize("newGame"), ui.TextConfig{}),
			OnSelect: func(menu *Menu) {
				menu.app.ProcessSignal(game.MapChangeSignal{
					NextMapPath: "assets/maps/e1m1.te3",
				})
			},
		},
		{
			Element: ui.NewText(ui.Transform{}, settings.Localize("loadGame"), ui.TextConfig{}),
			OnSelect: func(menu *Menu) {
				// TODO
			},
		},
		{
			Element: ui.NewText(ui.Transform{}, settings.Localize("saveGame"), ui.TextConfig{}),
			OnSelect: func(menu *Menu) {
				// TODO
			},
		},
		{
			Element: ui.NewText(ui.Transform{}, settings.Localize("options"), ui.TextConfig{}),
			OnSelect: func(menu *Menu) {
				// TODO
			},
		},
		{
			Element: ui.NewText(ui.Transform{}, settings.Localize("exit"), ui.TextConfig{}),
			OnSelect: func(menu *Menu) {
				os.Exit(0)
			},
		},
	}

	menuItems := titleScreenMenus[1:]
	if inGame {
		menuItems = titleScreenMenus[:]
		pos[1] -= 36.0
	}
	menu := newMenu(app, pos, menuItems)
	if inGame {
		menu.menuSelection = 0
	}
	return menu
}
