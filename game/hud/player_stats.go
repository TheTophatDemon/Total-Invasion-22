package hud

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	TEX_SEGAN_FACE = "assets/textures/ui/segan_face.png"
	TEX_HUD_ICONS  = "assets/textures/ui/hud_icons.png"
)

type PlayerStats struct {
	Health          int
	Noclip, GodMode bool
	Ammo            *game.Ammo
	Keys            game.KeyType
	MoveSpeed       float32
}

type faceState struct {
	anim     string
	flipX    bool
	showTime float32 // Number of seconds the face will appear for before giving way to lower priority states.
	priority int8
}

// Data about each face state.
var (
	FaceStateIdle      = faceState{anim: "idle", priority: 0}
	FaceStateHurtFront = faceState{anim: "hurt:front", showTime: 1.0, priority: 5}
	FaceStateHurtLeft  = faceState{anim: "hurt:side", flipX: true, showTime: 1.0, priority: 6}
	FaceStateHurtRight = faceState{anim: "hurt:side", flipX: false, showTime: 1.0, priority: 7}
	FaceStateDead      = faceState{anim: "dead", priority: 11}
	FaceStateNoclip    = faceState{anim: "noclip", priority: 10}
	FaceStateGod       = faceState{anim: "god", priority: 15}
)

var ammoTypeIconNames = [game.AMMO_TYPE_COUNT]string{
	game.AMMO_TYPE_NONE:   "",
	game.AMMO_TYPE_SICKLE: "sickle",
	game.AMMO_TYPE_EGG:    "egg",
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

	hudIconsTexture := cache.GetTexture(TEX_HUD_ICONS)

	// Heart icon
	var heart *ui.Box
	hud.heartIcon, heart, _ = hud.UI.Boxes.New()
	heartSlice := leftPanelTex.FindSlice("healthIcon")
	heart.SetTexture(hudIconsTexture).SetDest(fitToSlice(leftPanel.Dest(), heartSlice)).SetDepth(2.0)
	if heartAnim, ok := hudIconsTexture.GetAnimation("heart"); ok {
		heart.AnimPlayer.ChangeAnimation(heartAnim)
		heart.AnimPlayer.PlayFromStart()
	}

	// Face
	hud.faceState = FaceStateIdle
	faceTex := cache.GetTexture(TEX_SEGAN_FACE)
	var face *ui.Box
	hud.face, face, _ = hud.UI.Boxes.New()
	faceSlice := leftPanelTex.FindSlice("face")
	face.SetTexture(faceTex).SetDest(fitToSlice(leftPanel.Dest(), faceSlice)).SetDepth(2.0)
	if faceAnim, ok := faceTex.GetAnimation(FaceStateIdle.anim); ok {
		face.AnimPlayer.PlayNewAnim(faceAnim)
	}

	// Health counter
	var healthStat *ui.Text
	hud.healthStat, healthStat, _ = hud.UI.Texts.New()
	healthStatSlice := leftPanelTex.FindSlice("healthStat")
	healthStat.
		SetFont(COUNTER_FONT_PATH).
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

	// Ammo icon
	var ammoIcon *ui.Box
	hud.ammoIcon, ammoIcon, _ = hud.UI.Boxes.New()
	ammoIconSlice := rightPanelTex.FindSlice("ammoIcon")
	ammoIcon.SetTexture(hudIconsTexture).
		SetDest(fitToSlice(rightPanel.Dest(), ammoIconSlice)).
		SetDepth(2.0)
	ammoIcon.Hidden = true

	// Ammo counter
	var ammoStat *ui.Text
	hud.ammoStat, ammoStat, _ = hud.UI.Texts.New()
	ammoStatSlice := rightPanelTex.FindSlice("ammoStat")
	ammoStat.
		SetFont(COUNTER_FONT_PATH).
		SetText("000").
		SetDest(fitToSlice(rightPanel.Dest(), ammoStatSlice)).
		SetDepth(2.0).
		SetScale(SpriteScale()).
		SetAlignment(ui.TEXT_ALIGN_CENTER).
		SetColor(color.Blue)

	// Key icons
	for i, key := range [...]game.KeyType{game.KEY_TYPE_BLUE, game.KEY_TYPE_BROWN, game.KEY_TYPE_YELLOW, game.KEY_TYPE_GRAY} {
		var keyIcon *ui.Box
		hud.keyIcons[i], keyIcon, _ = hud.UI.Boxes.New()
		keyName := game.KeycardNames[key] + "Key"
		slice := rightPanelTex.FindSlice(keyName)
		keyIcon.SetDest(fitToSlice(rightPanel.Dest(), slice)).
			SetTexture(cache.GetTexture("assets/textures/ui/hud_keycards.png")).
			SetDepth(2.0)
		switch key {
		case game.KEY_TYPE_BLUE:
			keyIcon.SetSrc(math2.Rect{X: 0, Y: 0, Width: 8, Height: 8})
		case game.KEY_TYPE_BROWN:
			keyIcon.SetSrc(math2.Rect{X: 8, Y: 0, Width: 8, Height: 8})
		case game.KEY_TYPE_YELLOW:
			keyIcon.SetSrc(math2.Rect{X: 0, Y: 8, Width: 8, Height: 8})
		case game.KEY_TYPE_GRAY:
			keyIcon.SetSrc(math2.Rect{X: 8, Y: 8, Width: 8, Height: 8})
		}
		keyIcon.Hidden = true
	}
}

// Attempts to update the player's face to reflect in game events.
// If the current state has a higher priority and isn't out of time, then nothing will occur.
func (hud *Hud) SuggestPlayerFace(newState faceState) {
	if hud.faceState == newState || hud.faceState.priority > newState.priority {
		return
	}
	hud.forcePlayerFace(newState)
}

func (hud *Hud) forcePlayerFace(newState faceState) {
	if face, ok := hud.face.Get(); ok && hud.faceState != newState {
		faceTex := cache.GetTexture(TEX_SEGAN_FACE)
		anim, _ := faceTex.GetAnimation(newState.anim)
		face.AnimPlayer.PlayNewAnim(anim)
		face.FlippedHorz = newState.flipX
	}
	hud.faceState = newState
	hud.faceTimer = newState.showTime
}

func (hud *Hud) UpdatePlayerStats(deltaTime float32, stats PlayerStats) {
	// Health stat
	if txt, ok := hud.healthStat.Get(); ok {
		txt.SetText(fmt.Sprintf("%03d", stats.Health))
	}

	// Ammo stat
	if txt, ok := hud.ammoStat.Get(); ok {
		if weapon := hud.SelectedWeapon(); weapon != nil {
			txt.SetText(fmt.Sprintf("%03d", stats.Ammo[weapon.AmmoType()]))
			txt.Hidden = false
		} else {
			txt.Hidden = true
		}
	}
	if icon, ok := hud.ammoIcon.Get(); ok {
		if weapon := hud.SelectedWeapon(); weapon != nil {
			iconsTex := cache.GetTexture(TEX_HUD_ICONS)
			if anim, ok := iconsTex.GetAnimation(ammoTypeIconNames[weapon.AmmoType()]); ok && !icon.AnimPlayer.IsPlayingAnim(anim) {
				icon.AnimPlayer.PlayNewAnim(anim)
			}
			icon.Hidden = false
		} else {
			icon.Hidden = true
		}
	}

	// Keycards
	for i, keyHandle := range hud.keyIcons {
		if keySpr, ok := keyHandle.Get(); ok {
			keySpr.Hidden = (1<<i)&int(stats.Keys) == 0
		}
	}

	// Decide which face to display
	if stats.Health <= 0 {
		hud.forcePlayerFace(FaceStateDead)
		if heart, ok := hud.heartIcon.Get(); ok {
			heart.AnimPlayer.Stop()
		}
	} else if stats.GodMode {
		hud.forcePlayerFace(FaceStateGod)
	} else if stats.Noclip {
		hud.forcePlayerFace(FaceStateNoclip)
	} else {
		hud.faceTimer -= deltaTime
		if hud.faceTimer <= 0.0 {
			// Revert to idle face as a default
			hud.forcePlayerFace(FaceStateIdle)
		}
	}

	// Update weapons
	weapon := hud.Weapon(hud.selectedWeapon)
	if weapon == nil || !weapon.IsSelected() {
		hud.selectedWeapon = hud.nextWeapon
		if hud.selectedWeapon >= 0 {
			weapon = hud.Weapon(hud.selectedWeapon)
			weapon.Select()
		}
	}
	if weapon != nil && stats.Ammo != nil {
		weapon.Update(deltaTime, stats.MoveSpeed, stats.Ammo)
	}
}
