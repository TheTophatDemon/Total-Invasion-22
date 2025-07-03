package hud

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type PlayerStats struct {
	Health              int
	Noclip, GodMode     bool
	Ammo                *game.Ammo
	Armor               game.ArmorType
	ArmorAmount         int
	Keys                game.KeyType
	MoveSpeed           float32
	WeaponWheelOpenness float32 // Will be 1 when the weapon wheel is open and then gradually drop to 0 after the button is released.
}

type Hud struct {
	renderQueue ui.RenderQueue

	flashRect  ui.Box
	flashSpeed float32

	Debug         DebugStats
	Weapons       Weapons
	Intro         LevelIntro
	VictoryScreen VictoryScreen
	StatusBar     statusBar
	PlayerStats   PlayerStats // Information the player entity supplies
}

func (hud *Hud) Init() {
	hud.flashRect = ui.Box{
		Transform: ui.Transform{
			Dest: math2.Rect{
				Width:  settings.UIWidth(),
				Height: settings.UIHeight(),
			},
			Depth: 9.0,
			Scale: 1.0,
		},
		Color: color.Transparent,
	}

	if engine.InDebugMode() {
		hud.Debug.init()
	}
	hud.Weapons.init()
	hud.VictoryScreen.init()
	hud.StatusBar.init()
}

func (hud *Hud) Update(deltaTime float32) {
	// Update screen flash
	hud.flashRect.Color = hud.flashRect.Color.Fade(hud.flashSpeed * deltaTime)
	hud.renderQueue.Add(&hud.flashRect)

	if engine.InDebugMode() {
		hud.Debug.Layout(&hud.renderQueue)
	}
	if !hud.Intro.Done() {
		hud.Intro.Layout(&hud.renderQueue, deltaTime)
	}
	if hud.VictoryScreen.levelEndTime.IsZero() {
		hud.StatusBar.Layout(&hud.renderQueue, deltaTime, hud.PlayerStats, hud.Weapons.Selected())
		hud.Weapons.Layout(&hud.renderQueue, deltaTime, hud.PlayerStats)
	} else {
		// Only show after level ends.
		hud.VictoryScreen.Layout(&hud.renderQueue, deltaTime)
	}
}

func (hud *Hud) Render() {
	// Setup 2D render context
	renderContext := render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(0.0, float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight), 0.0, -50.0, 50.0),
	}
	hud.renderQueue.Render(&renderContext)
}

func (hud *Hud) ShowMessage(text string, duration float32, priority int, colr color.Color) {
	hud.StatusBar.ShowMessage(text, duration, priority, colr)
}

func (hud *Hud) FlashScreen(color color.Color, fadeSpeed float32) {
	hud.flashRect.Color = color
	hud.flashSpeed = fadeSpeed
}
