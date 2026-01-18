package screens

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type IntroScreen struct {
	title, flash, prompt  ui.Element
	app                   engine.Observer
	titleAnim             gween.Tween
	flashAnim, promptAnim gween.Sequence
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
	scr.titleAnim = *gween.New(0.0, scr.title.Size()[1], 1.0, ease.Linear) // Move title in from off screen
	scr.flashAnim = *gween.NewSequence(
		gween.New(0.0, 0.0, 0.5, ease.Linear),   // Wait
		gween.New(0.0, 1.0, 0.5, ease.InCubic),  // Flash in
		gween.New(1.0, 0.0, 1.0, ease.OutCubic), // Flash out
	)
	scr.promptAnim = *gween.NewSequence( // Blink on and off
		gween.New(1.0, 1.0, 0.75, ease.Linear),
		gween.New(0.0, 0.0, 0.25, ease.Linear),
	)
	scr.promptAnim.SetLoop(-1)

	return scr
}

func (scr *IntroScreen) Enter() {}
func (scr *IntroScreen) Exit()  {}

func (scr *IntroScreen) Layout(queue *ui.RenderQueue, deltaTime float32) {
	newY, _ := scr.titleAnim.Update(deltaTime)
	scr.title.SetY(newY)
	var complete bool
	scr.flash.BgColor.Unwrap().A, _, complete = scr.flashAnim.Update(deltaTime)
	queue.Add(&scr.flash)
	queue.Add(&scr.title)
	if complete {
		blink, _, _ := scr.promptAnim.Update(deltaTime)
		if blink > 0.0 {
			queue.Add(&scr.prompt)
		}
		if input.IsAnythingPressed() {
			scr.app.ProcessSignal(game.ChangeScreenSignal{
				Screen: new(TitleMenu).Init(scr.app, false),
			})
		}
	}
}
