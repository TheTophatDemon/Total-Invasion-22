package screens

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/engine/timer"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Chooser[T fmt.Stringer] struct {
	MenuItem
	txtLeft, txtValue, txtRight ui.Element
	labelWidth                  float32
	choices                     []T
	choice                      int
	blinkTimer                  timer.Timer
}

func (ch *Chooser[T]) Init(labelKey string, choices []T, choice int) *Chooser[T] {
	ch.MenuItem = MenuItem{
		element: ui.NewText(ui.Transform{}, settings.Localize(labelKey)+": ", ui.TextConfig{}),
	}
	ch.MenuItem.FitText()
	ch.txtLeft = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
		Size:   mgl32.Vec2{16, 24},
	}, "<", ui.TextConfig{})

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
	}, longestChoiceStr, ui.TextConfig{}) //TODO: TextAlignH breaks it??
	ch.txtValue.FitText()
	ch.txtValue.SetText(choices[choice].String())

	ch.txtRight = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
		Size:   mgl32.Vec2{16, 24},
	}, ">", ui.TextConfig{})
	ch.labelWidth = ch.OnScreenBox().Width
	ch.choices = choices
	ch.choice = choice
	ch.blinkTimer = timer.Timer{
		Interval: 0.5,
		Elapsed:  0.5,
	}
	return ch
}

func (ch *Chooser[T]) next() {
	ch.choice = (ch.choice + 1) % len(ch.choices)
}

func (ch *Chooser[T]) prev() {
	ch.choice = (ch.choice + len(ch.choices) - 1) % len(ch.choices)
}

func (ch *Chooser[T]) Input(action input.Action) {
	ch.MenuItem.Input(action)
	switch action {
	case settings.ActionMenuIncrement:
		ch.next()
	case settings.ActionMenuDecrement:
		ch.prev()
	case settings.ActionMenuClick:
		mousePos := input.MousePosition()
		if ch.txtLeft.OnScreenBox().ContainsPoint(mousePos) {
			ch.prev()
		} else if ch.txtRight.OnScreenBox().ContainsPoint(mousePos) {
			ch.next()
		}
	}
}

func (ch *Chooser[T]) Layout(queue *ui.RenderQueue, deltaTime float32) {
	ch.txtLeft.SetPosition(ch.Position())
	ch.txtLeft.Translate(mgl32.Vec2{ch.labelWidth, 0.0})
	queue.Add(&ch.txtLeft)

	leftWidth := ch.txtLeft.OnScreenBox().Width + 8.0
	ch.txtValue.SetPosition(ch.txtLeft.Position())
	ch.txtValue.SetText(ch.choices[ch.choice].String())
	ch.txtValue.Translate(mgl32.Vec2{leftWidth, 0.0})
	queue.Add(&ch.txtValue)

	valueWidth := ch.txtValue.OnScreenBox().Width + 8.0
	ch.txtRight.SetPosition(ch.txtValue.Position())
	ch.txtRight.Translate(mgl32.Vec2{valueWidth, 0.0})
	queue.Add(&ch.txtRight)

	fullWidth := ch.labelWidth + leftWidth + valueWidth + ch.txtRight.OnScreenBox().Width
	ch.SetSize(mgl32.Vec2{fullWidth, ch.Size()[1]})

	ch.MenuItem.Layout(queue, deltaTime)
}

func (ch *Chooser[T]) Choice() T {
	return ch.choices[ch.choice]
}
