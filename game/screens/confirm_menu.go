package screens

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/timer"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	// Confirms if the new screen resolution is suitable, or reverts it automatically after a timeout.
	ConfirmMenu struct {
		Menu
		timer            timer.Timer
		txtTimer         ui.Element
		labelItem        LabelItem
		confirmItem      MenuItem
		nextScreen       ui.Screen
		previousSettings settings.Data
	}
)

func (menu *ConfirmMenu) Init(app engine.Observer, parent *SettingsMenu, next ui.Screen, previousSettings settings.Data) *ConfirmMenu {
	*menu = ConfirmMenu{
		timer: timer.Timer{
			Interval: 1.0,
			MaxTicks: 3,
		},
		previousSettings: previousSettings,
		nextScreen:       next,
	}

	menu.confirmItem.Init("confirm", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		menu.app.ProcessSignal(game.ChangeScreenSignal{
			Screen: next,
		})
	})

	menu.labelItem.InitUnlocalized(fmt.Sprintf(settings.Localize("resizePrompt"), menu.timer.MaxTicks))

	menu.Menu.Init(
		app,
		[]MenuWidget{
			&menu.labelItem,
			&menu.confirmItem,
		},
		parent,
	)
	return menu
}

func (menu *ConfirmMenu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if menu.timer.Update(deltaTime) {
		if menu.timer.NumTicks == menu.timer.MaxTicks {
			settings.Current = menu.previousSettings
			settings.Current.Apply()
			menu.app.ProcessSignal(game.ChangeScreenSignal{
				Screen: new(SettingsMenu).Init(menu.app, menu.parent.(*SettingsMenu).parent),
			})
		} else {
			menu.labelItem.SetText(fmt.Sprintf(settings.Localize("resizePrompt"), menu.timer.MaxTicks-menu.timer.NumTicks))
		}
	}
	menu.Menu.Layout(queue, deltaTime)
}

func (menu *ConfirmMenu) Enter() {
}

func (menu *ConfirmMenu) Exit() {
}
