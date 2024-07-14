package hud

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	MESSAGE_FADE_SPEED = 2.0
	DEFAULT_FONT_PATH  = "assets/textures/ui/font.fnt"
)

type Hud struct {
	UI                        *ui.Scene
	FPSCounter, SpriteCounter scene.Id[*ui.Text]
	Heart, Face               scene.Id[*ui.Box]
	messageText               scene.Id[*ui.Text]
	messageTimer              float32
	messagePriority           int
	flashRect                 scene.Id[*ui.Box]
	flashSpeed                float32
}

func (hud *Hud) Init() {
	hud.UI = ui.NewUIScene(256, 64)

	hud.messageTimer = 2.0

	var fpsText *ui.Text
	hud.FPSCounter, fpsText, _ = hud.UI.Texts.New()
	fpsText.
		SetFont(DEFAULT_FONT_PATH).
		SetText("FPS: 0").
		SetDest(math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0}).
		SetScale(1.0).
		SetColor(color.White)

	var spriteCounter *ui.Text
	hud.SpriteCounter, spriteCounter, _ = hud.UI.Texts.New()
	spriteCounter.
		SetFont(DEFAULT_FONT_PATH).
		SetText("Sprites drawn: 0\nWalls drawn: 0\nParticles drawn: 0").
		SetDest(math2.Rect{X: 4.0, Y: 56.0, Width: 320.0, Height: 64.0}).
		SetScale(1.0).
		SetColor(color.Blue)

	var message *ui.Text
	hud.messageText, message, _ = hud.UI.Texts.New()
	message.
		SetFont(DEFAULT_FONT_PATH).
		SetText(settings.Localize("testMessage")).
		SetDest(math2.Rect{
			X:      settings.UIWidth() / 3.0,
			Y:      settings.UIHeight() / 2.0,
			Width:  settings.UIWidth() / 3.0,
			Height: settings.UIHeight() / 2.0,
		}).
		SetAlignment(ui.TEXT_ALIGN_CENTER).
		SetColor(color.Red).
		SetShadow(settings.Current.TextShadowColor, mgl32.Vec2{2.0, 2.0})

	var flashBox *ui.Box
	hud.flashRect, flashBox, _ = hud.UI.Boxes.New()
	flashBox.
		SetDest(math2.Rect{
			Width:  settings.UIWidth(),
			Height: settings.UIHeight(),
		}).
		SetDepth(9.0).
		SetColor(color.Blue.WithAlpha(0.5))

	hud.flashSpeed = 0.5

	// Left HUD panel
	leftPanelTex := cache.GetTexture("assets/textures/ui/hud_backdrop_left.png")
	_, leftPanel, _ := hud.UI.Boxes.New()
	panelHeight := float32(leftPanelTex.Height()) * SpriteScale()
	*leftPanel = ui.NewBoxFull(
		math2.Rect{
			X: 0.0, Y: settings.UIHeight() - panelHeight,
			Width:  float32(leftPanelTex.Width()) * SpriteScale(),
			Height: panelHeight,
		},
		leftPanelTex,
		color.White,
	)
	leftPanel.SetDepth(1.0)

	fitToSlice := func(parent math2.Rect, slice textures.Slice) math2.Rect {
		return math2.Rect{
			X:      parent.X + slice.Bounds.X*SpriteScale(),
			Y:      parent.Y + slice.Bounds.Y*SpriteScale(),
			Width:  slice.Bounds.Width * SpriteScale(),
			Height: slice.Bounds.Height * SpriteScale(),
		}
	}

	hudIconsTexture := cache.GetTexture("assets/textures/ui/hud_icons.png")

	// Heart icon
	var heart *ui.Box
	hud.Heart, heart, _ = hud.UI.Boxes.New()
	heartSlice := leftPanelTex.FindSlice("healthIcon")
	heart.SetTexture(hudIconsTexture).SetDest(fitToSlice(leftPanel.Dest(), heartSlice)).SetDepth(2.0)
	if heartAnim, ok := hudIconsTexture.GetAnimation("heart"); ok {
		heart.AnimPlayer.ChangeAnimation(heartAnim)
		heart.AnimPlayer.PlayFromStart()
	}

	// Face
	faceTex := cache.GetTexture("assets/textures/ui/segan_face.png")
	var face *ui.Box
	hud.Face, face, _ = hud.UI.Boxes.New()
	faceSlice := leftPanelTex.FindSlice("face")
	face.SetTexture(faceTex).SetDest(fitToSlice(leftPanel.Dest(), faceSlice)).SetDepth(2.0)
	if faceAnim, ok := faceTex.GetAnimation("idle"); ok {
		face.AnimPlayer.ChangeAnimation(faceAnim)
		face.AnimPlayer.PlayFromStart()
	}

	// Right HUD panel
	rightPanelTex := cache.GetTexture("assets/textures/ui/hud_backdrop_right.png")
	_, rightPanel, _ := hud.UI.Boxes.New()
	rightPanelWidth := rightPanelTex.Rect().Width * SpriteScale()
	*rightPanel = ui.NewBoxFull(
		math2.Rect{
			X:      settings.UIWidth() - rightPanelWidth,
			Y:      settings.UIHeight() - panelHeight,
			Width:  rightPanelWidth,
			Height: panelHeight,
		},
		rightPanelTex,
		color.White,
	)
	rightPanel.SetDepth(1.0)
}

func (hud *Hud) Update(deltaTime float32) {
	hud.UI.Update(deltaTime)

	// Update message text
	if message, ok := hud.messageText.Get(); ok {
		if hud.messageTimer > 0.0 {
			hud.messageTimer -= deltaTime
		} else {
			message.SetColor(message.Color().Fade(deltaTime * MESSAGE_FADE_SPEED))
			if message.Color().A <= 0.0 {
				message.SetColor(color.Transparent)
				hud.messageTimer = 0.0
				hud.messagePriority = 0
				message.SetText("")
			}
		}
	}

	// Update screen flash
	if flash, ok := hud.flashRect.Get(); ok {
		flash.Color = flash.Color.Fade(hud.flashSpeed * deltaTime)
	}

	// Update FPS counter
	if fpsText, ok := hud.FPSCounter.Get(); ok {
		fpsText.SetText(fmt.Sprintf("FPS: %v", engine.FPS()))
	}
}

func (hud *Hud) UpdateDebugCounters(renderContext *render.Context) {
	if sprCountTxt, ok := hud.SpriteCounter.Get(); ok {
		sprCountTxt.SetText(
			fmt.Sprintf("Sprites drawn: %v\nWalls drawn: %v\nParticles drawn: %v",
				renderContext.DrawnSpriteCount,
				renderContext.DrawnWallCount,
				renderContext.DrawnParticlesCount))
	}
}

func (hud *Hud) Render() {
	// Setup 2D render context
	renderContext := render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(0.0, float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight), 0.0, -10.0, 10.0),
	}

	// Render 2D game elements
	hud.UI.Render(&renderContext)
}

func (hud *Hud) ShowMessage(text string, duration float32, priority int, colr color.Color) {
	if priority >= hud.messagePriority {
		hud.messageTimer = duration
		hud.messagePriority = priority
		if message, ok := hud.messageText.Get(); ok {
			message.SetText(text).SetColor(colr)
		}
	}
}

func (hud *Hud) FlashScreen(color color.Color, fadeSpeed float32) {
	if flash, ok := hud.flashRect.Get(); ok {
		flash.Color = color
		hud.flashSpeed = fadeSpeed
	}
}

func SpriteScale() float32 {
	return settings.UIScale() * 2.0
}
