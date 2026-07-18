package screens

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const sfxChoose = "assets/sounds/ui/weapon_select.wav"

type Chooser[T fmt.Stringer] struct {
	MenuItem
	txtLabel, txtLeft, txtValue, txtRight ui.Element
	choices                               []T
	choiceIndex                           int
	startingChoice                        T
}

func (ch *Chooser[T]) Init(labelKey string, choices []T, choice T) *Chooser[T] {
	*ch = Chooser[T]{}

	smallScreen := settings.Current.WindowWidth <= 800

	ch.txtLabel = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
	}, settings.Localize(labelKey)+": ", ui.DefaultTextConfig())
	ch.txtLabel.FitText()
	if smallScreen && ch.txtLabel.Size()[0] >= 350.0 {
		ch.txtLabel.SetSize(mgl32.Vec2{
			350.0,
			48.0,
		})
	}

	ch.txtLeft = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
		Size:   mgl32.Vec2{16, 24},
	}, "<", ui.DefaultTextConfig().SetColor(color.Yellow).SetAlign(ui.TextAlignCenterV))

	longestChoiceStr := choices[0].String()
	for _, choice := range choices {
		str := choice.String()
		if len(str) > len(longestChoiceStr) {
			longestChoiceStr = str
		}
	}

	ch.txtValue = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
	}, longestChoiceStr, ui.DefaultTextConfig().SetAlign(ui.TextAlignCenterH|ui.TextAlignCenterV))
	ch.txtValue.FitText()
	ch.txtValue.SetText(choice.String())

	ch.txtRight = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
		Size:   mgl32.Vec2{16, 24},
	}, ">", ui.DefaultTextConfig().SetColor(color.Yellow).SetAlign(ui.TextAlignCenterV))
	ch.choices = choices
	ch.startingChoice = choice
	ch.choiceIndex = -1
	for i, existingChoice := range choices {
		if existingChoice.String() == choice.String() {
			ch.choiceIndex = i
			break
		}
	}
	return ch
}

func (ch *Chooser[T]) next() {
	cache.GetSfx(sfxChoose).Play()
	ch.choiceIndex = (ch.choiceIndex + 1) % len(ch.choices)
}

func (ch *Chooser[T]) prev() {
	cache.GetSfx(sfxChoose).Play()
	ch.choiceIndex = (ch.choiceIndex + len(ch.choices) - 1) % len(ch.choices)
}

func (ch *Chooser[T]) Input(action MenuInputType, menu *Menu) {
	ch.MenuItem.Input(action, menu)
	switch action {
	case MenuInputIncrement:
		ch.next()
	case MenuInputDecrement:
		ch.prev()
	default:
		mousePos := input.MousePosition()
		if ch.txtLeft.OnScreenBox().ContainsPoint(mousePos) {
			ch.prev()
		} else if ch.txtRight.OnScreenBox().ContainsPoint(mousePos) {
			ch.next()
		}
	}
}

func (ch *Chooser[T]) Focus() {
	ch.MenuItem.Focus()
	ch.SetTextColor(color.Yellow)
}

func (ch *Chooser[T]) Blur() {
	ch.MenuItem.Blur()
	ch.SetTextColor(color.White)
}

func (ch *Chooser[T]) Layout(queue *ui.RenderQueue, deltaTime float32) {

	ch.txtValue.SetText(ch.Choice().String())

	components := []*ui.Element{&ch.txtLabel, &ch.txtLeft, &ch.txtValue, &ch.txtRight}
	ui.LayoutStack(&ch.element, ui.StackParams{
		Gap: 8.0,
	}, components...)
	queue.Add(components...)

	ch.MenuItem.Layout(queue, deltaTime)
}

func (ch *Chooser[T]) Choice() T {
	if ch.choiceIndex >= 0 {
		return ch.choices[ch.choiceIndex]
	}
	return ch.startingChoice
}
