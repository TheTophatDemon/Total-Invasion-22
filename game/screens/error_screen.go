package screens

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type ErrorScreen struct {
	app        engine.Observer
	fade       ui.Element
	errorLabel ui.Element
}

func NewErrorScreen(app engine.Observer, err error) *ErrorScreen {
	scr := &ErrorScreen{
		app: app,
		errorLabel: ui.NewText(ui.Transform{
			Depth: 50,
			Size: mgl32.Vec2{
				settings.UIWidth(),
				settings.UIHeight(),
			},
		}, "Error\n"+err.Error(), ui.DefaultTextConfig()),
		fade: ui.NewBox(ui.Transform{
			Depth: 10,
			Size: mgl32.Vec2{
				settings.UIWidth(),
				settings.UIHeight(),
			},
		}, nil),
	}
	scr.fade.BgColor = maybe.Some(color.Black.WithAlpha(0.75))
	return scr
}

func (scr *ErrorScreen) Enter() {}
func (scr *ErrorScreen) Exit()  {}
func (scr *ErrorScreen) Bounds() math2.Rect {
	return math2.Rect{
		Width:  float32(settings.Current.WindowWidth),
		Height: float32(settings.Current.WindowHeight),
	}
}

func (scr *ErrorScreen) Layout(queue *ui.RenderQueue, deltaTime float32) {
	queue.Add(&scr.fade)
	queue.Add(&scr.errorLabel)
	if input.IsAnythingPressed() {
		scr.app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(TitleMenu).Init(scr.app, false),
		})
	}
}
