package menus

import (
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

var titleScreenMenus = [...]string{
	"Resume Game",
	"New Game",
	"Load Game",
	"Save Game",
	"Options",
	"Exit",
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
	menu.onSelect = menu.onTitleScreenSelect
	return menu
}

func (menu *Menu) onTitleScreenSelect(item string) {
	switch item {
	case titleScreenMenus[0]:
		menu.app.ProcessSignal(game.ResumeGameSignal{})
	case titleScreenMenus[1]:
		menu.app.ProcessSignal(game.MapChangeSignal{
			NextMapPath: "assets/maps/e1m1.te3",
		})
	case titleScreenMenus[5]:
		os.Exit(0)
	}
}
