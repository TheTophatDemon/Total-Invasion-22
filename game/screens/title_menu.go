package screens

import (
	"os"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/game"
)

type TitleMenu struct {
	Menu
	resumeGame, newGame, loadGame, saveGame, options, exit MenuItem
	inGame                                                 bool
}

func (titleMenu *TitleMenu) Init(app engine.Observer, inGame bool) *TitleMenu {
	*titleMenu = TitleMenu{
		inGame: inGame,
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
			Screen: new(SettingsMenu).Init(app, titleMenu),
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
	if !inGame {
		menuItems = menuItems[1:]
	}
	titleMenu.Menu.Init(app, menuItems, nil)
	if inGame {
		titleMenu.menuSelection = 0
	}
	return titleMenu
}
