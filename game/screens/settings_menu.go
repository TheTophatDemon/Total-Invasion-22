package screens

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	Resolution   [2]uint16
	SettingsMenu struct {
		Menu
		changedSettings   settings.Data
		returnItem        MenuItem
		chooserScreenSize Chooser[Resolution]
	}
)

func (res Resolution) String() string {
	return fmt.Sprintf("%vx%v", res[0], res[1])
}

func NewSettingsMenu(app engine.Observer, parent ui.Screen) *SettingsMenu {
	settingsMenu := &SettingsMenu{}
	settingsMenu.returnItem.Init("return", func(action input.Action) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: parent,
		})
	})
	settingsMenu.chooserScreenSize.Init("screenResolution", []Resolution{
		{640, 480},
		{800, 480},
		{800, 600},
		{1280, 720},
		{1920, 1080},
	}, 3)
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
	)
	return settingsMenu
}
