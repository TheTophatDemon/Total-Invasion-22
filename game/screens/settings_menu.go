package screens

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	Resolution   [2]uint16
	SettingsMenu struct {
		Menu
		returnItem        MenuItem
		chooserScreenSize Chooser[Resolution]
	}
)

func (res Resolution) String() string {
	return fmt.Sprintf("%vx%v", res[0], res[1])
}

func NewSettingsMenu(app engine.Observer, parent ui.Screen) *SettingsMenu {
	settingsMenu := &SettingsMenu{}
	settingsMenu.returnItem.Init("return", func(input.Action) { settingsMenu.returnToParent() })
	sizeChoices := []Resolution{
		{640, 480},
		{800, 480},
		{800, 600},
		{1280, 720},
		{1920, 1080},
	}
	sizeChoice := 3
	for i, choice := range sizeChoices {
		if choice[0] == uint16(settings.Current.WindowWidth) && choice[1] == uint16(settings.Current.WindowHeight) {
			sizeChoice = i
			break
		}
	}
	settingsMenu.chooserScreenSize.Init("screenResolution", sizeChoices, sizeChoice)
	settingsMenu.Menu = newMenu(
		app,
		mgl32.Vec2{
			(float32(settings.Current.WindowWidth) / 2.0) - 256.0,
			float32(settings.Current.WindowHeight) - 192.0,
		},
		[]MenuEvents{
			&settingsMenu.returnItem,
			&settingsMenu.chooserScreenSize,
		},
		parent,
	)
	return settingsMenu
}

func (menu *SettingsMenu) Exit() {
	newSize := menu.chooserScreenSize.Choice()
	if newSize[0] != settings.Current.WindowWidth || newSize[1] != settings.Current.WindowHeight {
		settings.Current.WindowWidth = newSize[0]
		settings.Current.WindowHeight = newSize[1]
		engine.SetScreenSize(int(settings.Current.WindowWidth), int(settings.Current.WindowHeight))
	}
}
