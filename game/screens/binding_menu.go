package screens

import (
	"fmt"
	"reflect"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type BindingMenu struct {
	Menu
	returnItem   MenuItem
	bindingItems []MenuItem
}

func (menu *BindingMenu) Init(app engine.Observer, parent ui.Screen) *BindingMenu {
	*menu = BindingMenu{}
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
		item.InitUnlocalized(actionString.String(), nil)
		menu.bindingItems = append(menu.bindingItems, item)
		allItems = append(allItems, &menu.bindingItems[len(menu.bindingItems)-1])
	}

	menu.Menu.Init(app, allItems, parent)

	return menu
}
