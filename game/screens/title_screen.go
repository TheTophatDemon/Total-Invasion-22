package screens

import (
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/game"
)

var titleScreenMenus = [...]string{
	"Resume Game",
	"New Game",
	"Load Game",
	"Save Game",
	"Options",
	"Exit",
}

func (scr *Screen) InitTitleScreen(app engine.Observer, position mgl32.Vec2, inGame bool) {
	menuItems := titleScreenMenus[1:]
	if inGame {
		menuItems = menuItems[:]
	}
	scr.init(app, position, menuItems)
	scr.onSelect = scr.onTitleScreenSelect
}

func (scr *Screen) onTitleScreenSelect(item string) bool {
	switch item {
	case titleScreenMenus[1]:
		scr.app.ProcessSignal(game.MapChangeSignal{
			NextMapPath: "assets/maps/e1m1.te3",
		})
		return false
	case titleScreenMenus[5]:
		os.Exit(0)
	}
	return true
}
