package hud

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	TEX_SEGAN_FACE   = "assets/textures/ui/segan_face.png"
	ANIM_FACE_IDLE   = "idle"
	ANIM_FACE_NOCLIP = "noclip"
)

type PlayerStats struct {
	Health int
	Noclip bool
}

func (hud *Hud) InitPlayerStats() {
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
	faceTex := cache.GetTexture(TEX_SEGAN_FACE)
	var face *ui.Box
	hud.Face, face, _ = hud.UI.Boxes.New()
	faceSlice := leftPanelTex.FindSlice("face")
	face.SetTexture(faceTex).SetDest(fitToSlice(leftPanel.Dest(), faceSlice)).SetDepth(2.0)
	if faceAnim, ok := faceTex.GetAnimation(ANIM_FACE_IDLE); ok {
		face.AnimPlayer.ChangeAnimation(faceAnim)
		face.AnimPlayer.PlayFromStart()
	}

	// Health counter
	var healthStat *ui.Text
	hud.healthStat, healthStat, _ = hud.UI.Texts.New()
	healthStatSlice := leftPanelTex.FindSlice("healthStat")
	healthStat.
		SetFont(COUTNER_FONT_PATH).
		SetText("000").
		SetDest(fitToSlice(leftPanel.Dest(), healthStatSlice)).
		SetDepth(2.0).
		SetScale(SpriteScale()).
		SetAlignment(ui.TEXT_ALIGN_CENTER).
		SetColor(color.Red)

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

func (hud *Hud) UpdatePlayerStats(deltaTime float32, stats PlayerStats) {
	if txt, ok := hud.healthStat.Get(); ok {
		txt.SetText(fmt.Sprintf("%03d", stats.Health))
	}

	if face, ok := hud.Face.Get(); ok {
		faceTex := cache.GetTexture(TEX_SEGAN_FACE)
		if stats.Noclip && face.AnimPlayer.CurrentAnimation().Name != ANIM_FACE_NOCLIP {
			anim, _ := faceTex.GetAnimation(ANIM_FACE_NOCLIP)
			face.AnimPlayer.ChangeAnimation(anim)
			face.AnimPlayer.PlayFromStart()
		} else if !stats.Noclip && face.AnimPlayer.CurrentAnimation().Name != ANIM_FACE_IDLE {
			anim, _ := faceTex.GetAnimation(ANIM_FACE_IDLE)
			face.AnimPlayer.ChangeAnimation(anim)
			face.AnimPlayer.PlayFromStart()
		}
	}
}
