package screens

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/engine/timer"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Chooser[T any] struct {
	MenuItem
	label      string
	choices    []T
	choice     int
	blinkTimer timer.Timer
}

func (ch *Chooser[T]) Init(labelKey string, choices []T, choice int) *Chooser[T] {
	ch.label = settings.Localize(labelKey)
	ch.MenuItem = MenuItem{
		element: ui.NewText(ui.Transform{}, ch.label, ui.TextConfig{}),
	}
	ch.choices = choices
	ch.choice = choice
	ch.blinkTimer = timer.Timer{
		Interval: 0.5,
		Elapsed:  0.5,
	}
	return ch
}

func (ch *Chooser[T]) Input(action input.Action) {
	ch.MenuItem.Input(action)
	switch action {
	case settings.ActionMenuIncrement:
		ch.choice = (ch.choice + 1) % len(ch.choices)
	case settings.ActionMenuDecrement:
		ch.choice = (ch.choice + len(ch.choices) - 1) % len(ch.choices)
	}
}

func (ch *Chooser[T]) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if ch.blinkTimer.Update(deltaTime) {
		format := "%v:   %v  "
		if (ch.blinkTimer.NumTicks % 2) == 1 {
			format = "%v: < %v >"
		}
		ch.element.SetText(fmt.Sprintf(format, ch.label, ch.choices[ch.choice]))
	}
	ch.MenuItem.Layout(queue, deltaTime)
}

func (ch *Chooser[T]) Choice() T {
	return ch.choices[ch.choice]
}
