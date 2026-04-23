package screens

import (
	"fmt"
	"reflect"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	BindingsMenu struct {
		Menu
	}
	actionItem struct {
		MenuItem
		bindingName string
		action      *settings.Action
	}
	bindingItem struct {
		MenuItem
		displayNumber int
		binding       *input.Binding
		clear         bool
	}
)

func (item *actionItem) Input(inputType MenuInputType, menu *Menu) {
	if item == nil || menu == nil || inputType == MenuInputDecrement || inputType == MenuInputIncrement {
		return
	}
	cache.GetSfx(SfxMenuHit).Play()
	menu.app.ProcessSignal(game.ChangeScreenSignal{
		Screen: new(BindingEditMenu).Init(menu.app, menu, item.action),
	})
}

func (item *actionItem) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if item == nil || item.action == nil {
		return
	}
	var actionString strings.Builder
	fmt.Fprintf(&actionString, "%v: ", item.bindingName)
	for i, binding := range item.action {
		if binding != nil {
			actionString.WriteString(settings.Localize(binding.LocalizationKey()))
		} else {
			actionString.WriteString("___")
		}
		if i != len(item.action)-1 {
			actionString.WriteString(", ")
		}
	}
	item.SetText(actionString.String())
	item.FitText()
	queue.Add(&item.element)
}

func (item *bindingItem) Input(inputType MenuInputType, menu *Menu) {
	if item == nil || inputType == MenuInputDecrement || inputType == MenuInputIncrement {
		return
	}
	cache.GetSfx(SfxMenuHit).Play()
	*item.binding = nil
	if !item.clear {
		input.CaptureInput(func(newBinding input.Binding, extraData any) {
			*item.binding = newBinding
		}, item)
	}
}

func (item *bindingItem) blankBindingString() string {
	return fmt.Sprintf(settings.Localize("setBindingFor"), item.displayNumber, "___")
}

func (item *bindingItem) captureBindingString() string {
	return fmt.Sprintf(settings.Localize("setBindingFor"), item.displayNumber, settings.Localize("inputPrompt"))
}

func (item *bindingItem) bindingString() string {
	return fmt.Sprintf(settings.Localize("setBindingFor"), item.displayNumber, settings.Localize((*item.binding).LocalizationKey()))
}

func (item *bindingItem) clearBindingString() string {
	return fmt.Sprintf(settings.Localize("clearBinding"), item.displayNumber)
}

func (item *bindingItem) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if item.clear {
		item.SetText(item.clearBindingString())
	} else if item.binding != nil && *item.binding != nil {
		item.SetText(item.bindingString())
	} else if input.CaptureExtraData() == item {
		item.SetText(item.captureBindingString())
	} else {
		item.SetText(item.blankBindingString())
	}
	item.FitText()
	queue.Add(&item.element)
}

func (menu *BindingsMenu) Init(app engine.Observer, parent ui.Screen) *BindingsMenu {
	*menu = BindingsMenu{}

	allItems := []MenuWidget{
		new(ReturnItem).Init(),
	}

	actionType := reflect.TypeFor[settings.Action]()
	for field, fieldValue := range reflect.ValueOf(&settings.Current).Elem().Fields() {
		if field.Type.Name() != actionType.Name() {
			continue
		}
		allItems = append(allItems, &actionItem{
			MenuItem: MenuItem{
				element: ui.NewText(ui.Transform{}, "?", ui.TextConfig{}),
			},
			bindingName: settings.Localize(strings.ToLower(string(field.Name[0])) + field.Name[1:]),
			action:      fieldValue.Addr().Interface().(*settings.Action),
		})
	}

	menu.Menu.Init(app, allItems, parent)

	return menu
}

type BindingEditMenu struct {
	Menu
}

func (menu *BindingEditMenu) Init(app engine.Observer, parent ui.Screen, action *settings.Action) *BindingEditMenu {
	if menu == nil || action == nil {
		return nil
	}
	*menu = BindingEditMenu{}
	allItems := []MenuWidget{
		new(ReturnItem).Init(),
	}
	for i := range action {
		displayNumber := i + 1
		item := &bindingItem{
			displayNumber: displayNumber,
			binding:       &action[i],
		}
		item.InitUnlocalized(item.captureBindingString(), nil)
		clearItem := &bindingItem{
			displayNumber: displayNumber,
			binding:       &action[i],
			clear:         true,
		}
		clearItem.InitUnlocalized(item.clearBindingString(), nil)
		allItems = append(allItems, item, clearItem)
	}
	menu.Menu.Init(app, allItems, parent)
	return menu
}
