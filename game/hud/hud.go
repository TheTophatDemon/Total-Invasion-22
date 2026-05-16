package hud

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tween"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

// Holds player information transferred between the HUD and the game worlds.
type PlayerStats struct {
	Health              int
	Noclip, GodMode     bool
	Ammo                game.Ammo
	Armor               game.ArmorType
	ArmorAmount         int
	Keys                game.Keys
	MoveSpeed           float32
	WeaponWheelOpenness float32 // Will be 1 when the weapon wheel is open and then gradually drop to 0 after the button is released.
	EquippedWeapons     [game.WeaponCount]bool
	SelectedWeapon      game.WeaponType
}

type Hud struct {
	renderQueue ui.RenderQueue

	flashBox  ui.Element
	flashAnim tween.Data

	frameBufferBox ui.Element // Contains the rendered game world

	Debug         DebugStats
	Weapons       Weapons
	Intro         LevelIntro
	VictoryScreen VictoryScreen
	StatusBar     statusBar
	MessageBar    MessageBar
	PlayerStats   PlayerStats // Information the player entity supplies
}

func (hud *Hud) Init() {
	hud.flashBox = ui.NewBox(ui.Transform{
		Size:  mgl32.Vec2{float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight)},
		Depth: 9.0,
	}, nil)
	hud.flashBox.BgColor = maybe.Some(color.Transparent)

	hud.frameBufferBox = ui.NewBox(ui.Transform{
		Size:  mgl32.Vec2{settings.UIWidth(), settings.UIHeight()},
		Depth: -1.0,
	}, nil)

	if engine.InDebugMode() {
		hud.Debug.init()
	}
	hud.Weapons.init(hud)
	hud.VictoryScreen.init()
	hud.StatusBar.init()
	hud.MessageBar.init()
}

func (hud *Hud) Update(deltaTime float32) {
	// Update screen flash
	if flashClr, ok := hud.flashBox.BgColor.Get(); ok {
		flashClr.A = hud.flashAnim.Update(deltaTime).Value
	}
	hud.renderQueue.Add(&hud.flashBox)

	if engine.InDebugMode() {
		hud.Debug.Layout(&hud.renderQueue)
	}
	if !hud.Intro.Done() {
		hud.Intro.Layout(&hud.renderQueue, deltaTime)
	}
	if hud.VictoryScreen.levelEndTime.IsZero() {
		hud.StatusBar.Layout(&hud.renderQueue, deltaTime, hud.PlayerStats, hud.Weapons.Selected())
		hud.MessageBar.layout(&hud.renderQueue, deltaTime)
		hud.Weapons.Layout(&hud.renderQueue, deltaTime, hud.PlayerStats)
	} else {
		// Only show after level ends.
		hud.VictoryScreen.Layout(&hud.renderQueue, deltaTime)
	}
}

func (hud *Hud) Render(frameBufferTexture *textures.Texture) {
	// Setup 2D render context
	renderContext := render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(0.0, float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight), 0.0, -50.0, 50.0),
	}
	renderContext.Enable2D()
	hud.frameBufferBox.BgTexture = frameBufferTexture
	hud.renderQueue.Add(&hud.frameBufferBox)
	hud.renderQueue.Render(&renderContext)
}

func (hud *Hud) ShowMessage(text string, priority int, colr color.Color) {
	hud.MessageBar.ShowMessage(text, priority, colr)
}

func (hud *Hud) FlashScreen(color color.Color, fadeDuration float32) {
	hud.flashBox.BgColor = maybe.Some(color)
	hud.flashAnim = tween.Data{
		StartValue: color.A,
		EndValue:   0,
		Duration:   fadeDuration,
	}
}

// Returns the size the sprites on the HUD should be scaled to.
func SpriteScale() float32 {
	return float32(settings.Current.WindowHeight) / 240
}
