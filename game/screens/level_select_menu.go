package screens

import (
	"io/fs"
	"path/filepath"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/game"
)

type (
	LevelSelectMenu struct {
		Menu
	}
)

func (lsm *LevelSelectMenu) Init(app engine.Observer) *LevelSelectMenu {
	*lsm = LevelSelectMenu{}

	menuItems := make([]MenuWidget, 0, 64)
	filepath.WalkDir("assets/maps", func(path string, d fs.DirEntry, err error) error {
		if strings.HasSuffix(path, ".te3") {
			fileName := filepath.Base(path[:len(path)-4])
			menuItems = append(menuItems, new(MenuItem).InitUnlocalized(fileName, func(menu *Menu, item MenuWidget, mit MenuInputType) {
				app.ProcessSignal(game.MapChangeSignal{
					MapPath: path,
				})
			}))
		}
		return nil
	})

	lsm.Menu.Init(app, menuItems, nil)
	return lsm
}
