package screens

import (
	"os"

	"github.com/go-gl/mathgl/mgl32"
)

var titleScreenMenus = [...]string{
	"Resume Game",
	"New Game",
	"Load Game",
	"Save Game",
	"Options",
	"Exit",
}

func (scr *Screen) InitTitleScreen(position mgl32.Vec2, inGame bool) {
	menuItems := titleScreenMenus[1:]
	if inGame {
		menuItems = menuItems[:]
	}
	scr.init(position, menuItems)
	scr.onSelect = scr.onTitleScreenSelect
}

func (scr *Screen) onTitleScreenSelect(item string) {
	switch item {
	case titleScreenMenus[5]:
		os.Exit(0)
	}
}
