package screens

import (
	"fmt"
	"reflect"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type BindingsMenu struct {
	Menu
	returnItem  MenuItem
	actionItems []MenuItem
}

func (menu *BindingsMenu) Init(app engine.Observer, parent ui.Screen) *BindingsMenu {
	*menu = BindingsMenu{}
	menu.returnItem.Init("return", func(MenuInputType) { menu.returnToParent() })

	allItems := []MenuEvents{
		&menu.returnItem,
	}

	actionType := reflect.TypeFor[settings.Action]()
	for field, fieldValue := range reflect.ValueOf(settings.Current).Fields() {
		if field.Type.Name() != actionType.Name() {
			continue
		}
		item := MenuItem{}
		bindingName := settings.Localize(strings.ToLower(string(field.Name[0])) + field.Name[1:])
		action := fieldValue.Interface().(settings.Action)
		var actionString strings.Builder
		fmt.Fprintf(&actionString, "%v: ", bindingName)
		for i, binding := range action {
			if binding != nil {
				actionString.WriteString(settings.Localize(binding.LocalizationKey()))
			} else {
				actionString.WriteString("___")
			}
			if i != len(action)-1 {
				actionString.WriteString(", ")
			}
		}
		item.InitUnlocalized(actionString.String(), func(MenuInputType) {
			app.ProcessSignal(game.ChangeScreenSignal{
				Screen: new(BindingEditMenu).Init(app, menu, &action),
			})
		})
		menu.actionItems = append(menu.actionItems, item)
		allItems = append(allItems, &menu.actionItems[len(menu.actionItems)-1])
	}

	menu.Menu.Init(app, allItems, parent)

	return menu
}

type BindingEditMenu struct {
	Menu
	returnItem   MenuItem
	bindingItems []MenuItem
	action       *settings.Action
	editedAction settings.Action
}

func (menu *BindingEditMenu) Init(app engine.Observer, parent ui.Screen, action *settings.Action) *BindingEditMenu {
	if menu == nil || action == nil {
		return nil
	}
	*menu = BindingEditMenu{
		action:       action,
		editedAction: *action,
	}
	menu.returnItem.Init("return", func(MenuInputType) { menu.returnToParent() })
	allItems := []MenuEvents{
		&menu.returnItem,
	}
	const blankBindingText = "___"
	for i, binding := range action {
		item := MenuItem{}
		var bindingString strings.Builder
		bindingName := blankBindingText
		if binding != nil {
			bindingName = settings.Localize(binding.LocalizationKey())
		}
		displayNumber := i + 1
		fmt.Fprintf(&bindingString, settings.Localize("setBindingFor"), displayNumber, bindingName)
		item.InitUnlocalized(bindingString.String(), func(mit MenuInputType) {
			item.SetText(fmt.Sprintf(settings.Localize("setBindingFor"), displayNumber, settings.Localize("inputPrompt")))
			item.FitText()
			input.CaptureInput(func(newBinding input.Binding) {
				item.SetText(fmt.Sprintf(settings.Localize("setBindingFor"), displayNumber, settings.Localize(newBinding.LocalizationKey())))
				item.FitText()
				action[i] = newBinding
			})
		})
		clearItem := MenuItem{}
		clearItem.InitUnlocalized(fmt.Sprintf(settings.Localize("clearBinding"), displayNumber), func(MenuInputType) {
			item.SetText(fmt.Sprintf(settings.Localize("setBindingFor"), displayNumber, blankBindingText))
			item.FitText()
			action[i] = nil
		})
		allItems = append(allItems, &item, &clearItem)
	}
	menu.Menu.Init(app, allItems, parent)
	return menu
}
