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

type (
	SaveMenu struct {
		Menu
		currentPreview, filePreview SavePreview
		previewTimer                timer.Timer
	}
	SaveGameItem struct {
		MenuItem
		SaveData   game.MapChangeSignal
		SaveNumber int
		menu       *SaveMenu
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
	item.menu.filePreview.SaveData = &item.SaveData
	item.menu.previewTimer = timer.Timer{}
}

func (item *SaveGameItem) Blur() {
	item.MenuItem.Blur()
	item.menu.previewTimer = timer.Timer{
		Interval: 1.0,
		MaxTicks: 1,
	}
}

func handleSaveClick(menu *Menu, item MenuWidget, mit MenuInputType) {
	saveItem := item.(*SaveGameItem)
	saveItem.SetText(settings.Localize("fileSaved"))
	saveItem.SaveData = *saveItem.menu.currentPreview.SaveData
	menu.app.ProcessSignal(game.SaveSignal{
		Number:   saveItem.SaveNumber,
		WithData: &saveItem.SaveData,
	})
}

func (sm *SaveMenu) Init(app engine.Observer, parent ui.Screen) *SaveMenu {
	*sm = SaveMenu{}

	sm.filePreview.Init("fileProgress")
	sm.currentPreview.Init("yourProgress")

	menuItems := make([]MenuWidget, 1, SaveFileCount+1)
	menuItems[0] = new(ReturnItem).Init()
	for i := range SaveFileCount {
		var menuItem *SaveGameItem
		var saveData game.MapChangeSignal
		if i != 0 {
			menuItem = new(SaveGameItem)
			saveName := fmt.Sprintf(settings.Localize("copyToFile"), i)
			menuItem.InitUnlocalized(saveName, handleSaveClick)
			menuItem.menu = sm
			menuItem.SaveNumber = i
		}

		saveFile, err := os.Open(fmt.Sprintf("save%d", i))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				failure.LogErrWithLocation("failed to open save file %d: %d", i, err)
			}
		} else {
			defer saveFile.Close()
			saveBytes, err := io.ReadAll(saveFile)
			if err != nil {
				failure.LogErrWithLocation("failed to read from save file %d: %d", i, err)
				continue
			}

			err = json.Unmarshal(saveBytes, &saveData)
			if err != nil {
				failure.LogErrWithLocation("failed to parse from save file %d: %d", i, err)
				continue
			}
		}

		if i == 0 {
			sm.currentPreview.SaveData = &saveData
		} else {
			menuItem.SaveData = saveData
			menuItems = append(menuItems, menuItem)
		}
	}

	sm.Menu.Init(app, menuItems, parent)
	return sm
}

func (sm *SaveMenu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	sm.Menu.Layout(queue, deltaTime)
	if sm.currentPreview.SaveData != nil {
		menuBounds := sm.Menu.background.OnScreenBox()
		sm.currentPreview.SetPosition(menuBounds.PosVec().Add(mgl32.Vec2{menuBounds.Width + 8, 0}))
		sm.currentPreview.SetSize(mgl32.Vec2{settings.UIWidth() - 16.0 - sm.currentPreview.Position()[0], (menuBounds.Height / 2.0) - 8})
		sm.currentPreview.Layout(queue, deltaTime)

		sm.filePreview.SetPosition(sm.currentPreview.Position().Add(mgl32.Vec2{0, sm.currentPreview.Height() + 8}))
		sm.filePreview.SetSize(sm.currentPreview.Size())
		sm.filePreview.Layout(queue, deltaTime)
		if sm.previewTimer.Update(deltaTime) {
			sm.filePreview.SaveData = nil
		}
	}
}
