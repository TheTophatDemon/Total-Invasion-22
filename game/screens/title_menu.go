package screens

import (
	"tophatdemon.com/total-invasion-ii/engine"
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
	titleMenu.resumeGame.Init("resumeGame", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		app.ProcessSignal(game.ResumeGameSignal{})
	})
	titleMenu.newGame.Init("newGame", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(DifficultyMenu).Init(app, titleMenu),
		})
	})
	titleMenu.loadGame.Init("loadGame", func(m *Menu, mw MenuWidget, mit MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(LoadMenu).Init(app, titleMenu),
		})
	})
	titleMenu.saveGame.Init("saveGame", func(m *Menu, mw MenuWidget, mit MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(SaveMenu).Init(app, titleMenu),
		})
	})
	titleMenu.options.Init("options", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(SettingsMenu).Init(app, titleMenu),
		})
	})
	titleMenu.exit.Init("exit", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		engine.Shutdown()
	})
	menuItems := []MenuWidget{
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
	return titleMenu
}
