package screens

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

var titleScreenMenus = [...]string{
	"Resume Game",
	"New Game",
	"Load Game",
	"Save Game",
	"Options",
	"Exit",
}

type TitleScreen struct {
	menuSelection int
}

func (ts *TitleScreen) Title() string {
	return "Title Screen"
}

func (ts *TitleScreen) Layout(queue *ui.RenderQueue, position mgl32.Vec2, deltaTime float32) {
	if input.IsActionJustPressed(settings.ActionMenuDown) {
		ts.menuSelection++
	} else if input.IsActionJustPressed(settings.ActionMenuUp) {
		ts.menuSelection--
	}
	if ts.menuSelection >= len(titleScreenMenus) {
		ts.menuSelection -= len(titleScreenMenus)
	}
	if ts.menuSelection < 1 {
		ts.menuSelection = 1
	}

	for i, text := range titleScreenMenus {
		if i == 0 {
			// TODO: Resume game
			continue
		}
		elem := &ui.Text{
			Settings: ui.TextSettings{
				Text: text,
			},
			Transform: ui.Transform{
				Dest: math2.Rect{
					X:      position[0],
					Y:      position[1] + (16.0 * float32(i)),
					Width:  settings.UIWidth(),
					Height: 16.0,
				},
			},
		}

		if ts.menuSelection == i {
			elem.Color = color.Yellow
		}

		queue.Add(elem)
	}
}
