package screens

import (
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type TitleMenu struct {
	Menu
	resumeGame, newGame, loadGame, saveGame, options, exit MenuItem
}

func NewTitleMenu(app engine.Observer, inGame bool) *TitleMenu {
	titleMenu := &TitleMenu{}
	pos := mgl32.Vec2{
		(float32(settings.Current.WindowWidth) / 2.0) - 256.0,
		float32(settings.Current.WindowHeight) - 192.0,
	}
	titleMenu.resumeGame.Init("resumeGame", func(input.Action) {
		app.ProcessSignal(game.ResumeGameSignal{})
	})
	titleMenu.newGame.Init("newGame", func(input.Action) {
		app.ProcessSignal(game.MapChangeSignal{
			NextMapPath: "assets/maps/e1m1.te3",
		})
	})
	titleMenu.loadGame.Init("loadGame", nil)
	titleMenu.saveGame.Init("saveGame", nil)
	titleMenu.options.Init("options", func(input.Action) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: NewSettingsMenu(app, titleMenu),
		})
	})
	titleMenu.exit.Init("exit", func(input.Action) {
		os.Exit(0)
	})
	menuItems := []MenuEvents{
		&titleMenu.resumeGame,
		&titleMenu.newGame,
		&titleMenu.loadGame,
		&titleMenu.saveGame,
		&titleMenu.options,
		&titleMenu.exit,
	}
	if inGame {
		pos[1] -= 36.0
	} else {
		menuItems = menuItems[1:]
	}
	titleMenu.Menu = newMenu(app, pos, menuItems)
	if inGame {
		titleMenu.menuSelection = 0
	}
	return titleMenu
}
