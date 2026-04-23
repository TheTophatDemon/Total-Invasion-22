package screens

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tween"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type IntroScreen struct {
	title, flash, prompt  ui.Element
	app                   engine.Observer
	titleAnim             tween.Data
	flashAnim, promptAnim tween.Sequence
}

func NewIntroScreen(app engine.Observer) *IntroScreen {
	scr := &IntroScreen{
		app: app,
		title: ui.NewBox(ui.Transform{
			Anchor:   ui.Ratios{0.5, 0.0},
			Origin:   ui.Ratios{0.5, 1.0},
			Position: mgl32.Vec2{0.0, 0.0},
			Depth:    50,
		}, cache.GetTexture("assets/textures/ui/title.png")),
		flash: ui.NewBox(ui.Transform{
			Size:  mgl32.Vec2{float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight)},
			Depth: 1,
		}, nil),
		prompt: ui.NewText(
			ui.Transform{
				Size:     mgl32.Vec2{512.0, 64.0},
				Anchor:   ui.Ratios{0.5, 1.0},
				Origin:   ui.Ratios{0.5, 1.0},
				Position: mgl32.Vec2{0.0, -48.0},
			},
			settings.Localize("startPrompt"),
			ui.TextConfig{
				Align: ui.TextAlignCenterH,
				Color: maybe.Some(color.Yellow),
			},
		),
	}
	scr.flash.BgColor = maybe.Some(color.White.WithAlpha(0.0))
	scr.title.FitHeight(float32(settings.Current.WindowHeight) * 0.355555556)

	// Move title in from off screen
	scr.titleAnim = tween.Data{
		EndValue: scr.title.Height(),
		Duration: 1.0,
	}

	scr.flashAnim = tween.Sequence{
		Tweens: []tween.Data{
			{EndValue: 0.8, Duration: 0.75, Interpolation: tween.CubicIn},
			{StartValue: tween.Infer, EndValue: 0.0, Duration: 0.75, Interpolation: tween.CubicOut},
		},
	}

	// Blink on and off
	scr.promptAnim = tween.Sequence{
		Tweens: []tween.Data{
			{StartValue: 1.0, EndValue: tween.Infer, Duration: 0.75},
			{Duration: 0.25},
		},
		Loop: true,
	}

	return scr
}

func (scr *IntroScreen) Enter() {}
func (scr *IntroScreen) Exit()  {}
func (scr *IntroScreen) Bounds() math2.Rect {
	return math2.Rect{
		Width:  float32(settings.Current.WindowWidth),
		Height: float32(settings.Current.WindowHeight),
	}
}

func (scr *IntroScreen) Layout(queue *ui.RenderQueue, deltaTime float32) {
	scr.title.SetY(scr.titleAnim.Update(deltaTime).Value)
	flashRes := scr.flashAnim.Update(deltaTime)
	scr.flash.BgColor.Unwrap().A = flashRes.Value
	queue.Add(&scr.flash)
	queue.Add(&scr.title)
	if flashRes.SequenceDone {
		if scr.promptAnim.Update(deltaTime).Value > 0.0 {
			queue.Add(&scr.prompt)
		}
		if input.IsAnythingPressed() {
			scr.app.ProcessSignal(game.ChangeScreenSignal{
				Screen: new(TitleMenu).Init(scr.app, false),
			})
		}
	}
}
