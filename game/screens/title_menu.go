package screens

import (
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

var titleScreenMenus = [...]string{
	"resumeGame",
	"newGame",
	"loadGame",
	"saveGame",
	"options",
	"exit",
}

func NewTitleMenu(app engine.Observer, inGame bool) *Menu {
	pos := mgl32.Vec2{
		(float32(settings.Current.WindowWidth) / 2.0) - 256.0,
		float32(settings.Current.WindowHeight) - 192.0,
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
	menu.onSelect = menu.onTitleScreenSelect
	return menu
}

func (menu *Menu) onTitleScreenSelect(itemIndex int) {
	switch menu.menuTexts[itemIndex].Text() {
	case settings.Localize("resumeGame"):
		menu.app.ProcessSignal(game.ResumeGameSignal{})
	case settings.Localize("newGame"):
		menu.app.ProcessSignal(game.MapChangeSignal{
			NextMapPath: "assets/maps/e1m1.te3",
		})
	case settings.Localize("exit"):
		os.Exit(0)
	}
}
