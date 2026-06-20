package screens

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const SaveFileCount = 10

type (
	LoadMenu struct {
		Menu
	}
	SaveGameItem struct {
		MenuItem
		SaveData game.MapChangeSignal
	}
)

func (item *SaveGameItem) Input(action MenuInputType, menu *Menu) {
	if (action == MenuInputConfirm || action == MenuInputClick) && item.OnInput != nil {
		item.OnInput(menu, item, action)
		cache.GetSfx(SfxMenuHit).Play()
	}
}

func handleLoadClick(menu *Menu, item MenuWidget, mit MenuInputType) {
	saveItem := item.(*SaveGameItem)
	menu.app.ProcessSignal(saveItem.SaveData)
}

func (lm *LoadMenu) Init(app engine.Observer, parent ui.Screen) *LoadMenu {
	*lm = LoadMenu{}

	menuItems := make([]MenuWidget, SaveFileCount+1)
	menuItems[0] = new(ReturnItem).Init()
	for i := range SaveFileCount {
		var saveName string
		if i == 0 {
			saveName = settings.Localize("autoSaveFile")
		} else {
			saveName = fmt.Sprintf(settings.Localize("saveFile"), i)
		}
		menuItem := new(SaveGameItem)
		menuItem.InitUnlocalized(saveName, handleLoadClick)
		menuItems[i+1] = menuItem

		saveFile, err := os.Open(fmt.Sprintf("save%d", i))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				failure.LogErrWithLocation("failed to open save file %d: %d", i, err)
			}
			continue
		}
		saveBytes, err := io.ReadAll(saveFile)
		if err != nil {
			failure.LogErrWithLocation("failed to read from save file %d: %d", i, err)
			continue
		}
		err = json.Unmarshal(saveBytes, &menuItem.SaveData)
		if err != nil {
			failure.LogErrWithLocation("failed to parse from save file %d: %d", i, err)
			continue
		}

		defer saveFile.Close()
	}

	lm.Menu.Init(app, menuItems, parent)
	return lm
}
