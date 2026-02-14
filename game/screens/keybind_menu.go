package screens

import (
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
)

type KeybindMenu struct {
	Menu
	returnItem MenuItem
}

func (keybindMenu *KeybindMenu) Init(app engine.Observer, parent ui.Screen) *KeybindMenu {
	*keybindMenu = KeybindMenu{}
	keybindMenu.returnItem.Init("return", func(MenuInputType) { keybindMenu.returnToParent() })

	keybindMenu.Menu.Init(app, []MenuEvents{
		&keybindMenu.returnItem,
	}, parent)
	return keybindMenu
}
