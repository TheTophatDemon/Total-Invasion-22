package screens

import (
	"fmt"
	"os"

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

	menuItems := make([]MenuWidget, 0, 6)

	if inGame {
		titleMenu.resumeGame.Init("resumeGame", func(menu *Menu, item MenuWidget, mit MenuInputType) {
			app.ProcessSignal(game.ResumeGameSignal{})
		})
		menuItems = append(menuItems, &titleMenu.resumeGame)
	}

	titleMenu.newGame.Init("newGame", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(DifficultyMenu).Init(app, titleMenu),
		})
	})
	menuItems = append(menuItems, &titleMenu.newGame)

	// Count the number of save files available to see if we need the save or load menus
	var hasAnySaveFile, hasAutoSaveFile bool
	for i := range SaveFileCount {
		_, err := os.Stat(fmt.Sprintf("save%v", i))
		if err == nil {
			if i == 0 {
				hasAutoSaveFile = true
			}
			hasAnySaveFile = true
			break
		}
	}

	if hasAnySaveFile {
		titleMenu.loadGame.Init("loadGame", func(m *Menu, mw MenuWidget, mit MenuInputType) {
			app.ProcessSignal(game.ChangeScreenSignal{
				Screen: new(LoadMenu).Init(app, titleMenu),
			})
		})
		menuItems = append(menuItems, &titleMenu.loadGame)
	}

	if hasAutoSaveFile {
		titleMenu.saveGame.Init("saveGame", func(m *Menu, mw MenuWidget, mit MenuInputType) {
			app.ProcessSignal(game.ChangeScreenSignal{
				Screen: new(SaveMenu).Init(app, titleMenu),
			})
		})
		menuItems = append(menuItems, &titleMenu.saveGame)
	}

	titleMenu.options.Init("options", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(SettingsMenu).Init(app, titleMenu),
		})
	})
	menuItems = append(menuItems, &titleMenu.options)

	titleMenu.exit.Init("exit", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		engine.Shutdown()
	})
	menuItems = append(menuItems, &titleMenu.exit)

	titleMenu.Menu.Init(app, menuItems, nil)
	return titleMenu
}
