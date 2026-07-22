package screens

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/timer"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const SaveFileCount = 10

type (
	LoadMenu struct {
		Menu
		preview      SavePreview
		previewTimer timer.Timer
	}
	SaveGameItem struct {
		MenuItem
		SaveData game.MapChangeSignal
		menu     *LoadMenu
	}
)

func (item *SaveGameItem) Input(action MenuInputType, menu *Menu) {
	if (action == MenuInputConfirm || action == MenuInputClick) && item.OnInput != nil {
		item.OnInput(menu, item, action)
		cache.GetSfx(SfxMenuHit).Play()
	}
}

func (item *SaveGameItem) Focus() {
	item.MenuItem.Focus()
	item.menu.preview.SaveData = &item.SaveData
	item.menu.previewTimer = timer.Timer{}
}

func (item *SaveGameItem) Blur() {
	item.MenuItem.Blur()
	item.menu.previewTimer = timer.Timer{
		Interval: 1.0,
		MaxTicks: 1,
	}
}

func handleLoadClick(menu *Menu, item MenuWidget, mit MenuInputType) {
	saveItem := item.(*SaveGameItem)
	menu.app.ProcessSignal(saveItem.SaveData)
}

func (lm *LoadMenu) Init(app engine.Observer, parent ui.Screen) *LoadMenu {
	*lm = LoadMenu{}

	lm.preview.Init()

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
		menuItem.menu = lm
		menuItems[i+1] = menuItem

		saveFile, err := os.Open(fmt.Sprintf("save%d", i))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				failure.LogErrWithLocation("failed to open save file %d: %d", i, err)
			}
			continue
		}
		defer saveFile.Close()

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
	}

	lm.Menu.Init(app, menuItems, parent)
	return lm
}

func (lm *LoadMenu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	lm.Menu.Layout(queue, deltaTime)
	if lm.preview.SaveData != nil {
		menuBounds := lm.Menu.background.OnScreenBox()
		lm.preview.SetPosition(menuBounds.PosVec().Add(mgl32.Vec2{menuBounds.Width + 8, 0}))
		lm.preview.SetSize(mgl32.Vec2{settings.UIWidth() - 16.0 - lm.preview.Position()[0], menuBounds.Height})
		lm.preview.Layout(queue, deltaTime)
		if lm.previewTimer.Update(deltaTime) {
			lm.preview.SaveData = nil
		}
	}
}
