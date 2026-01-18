package screens

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	Resolution   [2]uint16
	OnOff        bool
	SettingsMenu struct {
		Menu
		returnItem        MenuItem
		chooserScreenSize Chooser[Resolution]
		chooserFullscreen Chooser[OnOff]
	}
)

func (res Resolution) String() string {
	return fmt.Sprintf("%vx%v", res[0], res[1])
}

func (onOff OnOff) String() string {
	if onOff {
		return settings.Localize("on")
	}
	return settings.Localize("off")
}

func (settingsMenu *SettingsMenu) Init(app engine.Observer, parent ui.Screen) *SettingsMenu {
	*settingsMenu = SettingsMenu{}
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

	fullscreenChoice := 0
	if settings.Current.Fullscreen {
		fullscreenChoice = 1
	}
	settingsMenu.chooserFullscreen.Init("fullscreen", []OnOff{OnOff(false), OnOff(true)}, fullscreenChoice)
	settingsMenu.Menu.Init(
		app,
		[]MenuEvents{
			&settingsMenu.returnItem,
			&settingsMenu.chooserScreenSize,
			&settingsMenu.chooserFullscreen,
		},
		parent,
	)
	return settingsMenu
}

func (menu *SettingsMenu) Exit() {
	originalSettings := settings.Current
	resetMenu := false
	newSize := menu.chooserScreenSize.Choice()
	if nowFullscreen := bool(menu.chooserFullscreen.Choice()); nowFullscreen != engine.IsFullscreen() {
		settings.Current.Fullscreen = nowFullscreen
		engine.SetFullscreen(nowFullscreen)
		resetMenu = true
	}
	if newSize[0] != settings.Current.WindowWidth || newSize[1] != settings.Current.WindowHeight {
		settings.Current.WindowWidth = newSize[0]
		settings.Current.WindowHeight = newSize[1]
		engine.SetScreenSize(int(settings.Current.WindowWidth), int(settings.Current.WindowHeight))
		resetMenu = true
	}
	if resetMenu {
		// Reset the title menu so things are positioned correctly
		if titleMenu, ok := menu.parent.(*TitleMenu); ok {
			titleMenu.Init(titleMenu.app, titleMenu.inGame)
		}
	}
	if originalSettings != settings.Current {
		settings.Save()
	}
}
