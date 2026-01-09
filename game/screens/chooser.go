package screens

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/engine/timer"
)

type Chooser[T any] struct {
	ui.Element
	choices    []T
	choice     int
	textConfig ui.TextConfig
	blinkTimer timer.Timer
}

func (ch *Chooser[T]) Init(choices []T, choice int, transform ui.Transform, config ui.TextConfig) *Chooser[T] {
	ch.Element = ui.NewText(transform, "", config)
	ch.choices = choices
	ch.choice = choice
	ch.textConfig = config
	ch.blinkTimer = timer.Timer{
		Interval: 0.5,
	}
	return ch
}

func (ch *Chooser[T]) Update(deltaTime float32) {
	ch.choice = max(min(ch.choice, len(ch.choices)-1), 0)
	if ch.blinkTimer.Update(deltaTime) {
		format := "  %v  "
		if (ch.blinkTimer.NumTicks % 2) == 1 {
			format = "< %v >"
		}
		ch.SetText(fmt.Sprintf(format, ch.choices[ch.choice]))
	}
}

func (ch *Chooser[T]) Choice() T {
	return ch.choices[ch.choice]
}
