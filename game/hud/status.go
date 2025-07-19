package hud

import (
	"fmt"
	"log"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	texSeganFace = "assets/textures/ui/segan_face.png"
	texHudIcons  = "assets/textures/ui/hud_icons.png"
)

type statusBar struct {
	leftPanel, rightPanel           ui.Box
	face                            ui.Box
	faceState                       faceState
	faceTimer                       float32
	heartIcon, ammoIcon, armorIcon  ui.Box
	keyIcons                        [4]ui.Box
	healthStat, ammoStat, armorStat ui.Text
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

var ammoTypeIconNames = [game.AmmoTypeCount]string{
	game.AmmoTypeNone:    "",
	game.AmmoTypeSickle:  "sickle",
	game.AmmoTypeEgg:     "egg",
	game.AmmoTypeGrenade: "grenade",
	game.AmmoTypePlasma:  "plasma",
}

// Attempts to update the player's face to reflect in game events.
// If the current state has a higher priority and isn't out of time, then nothing will occur.
func (status *statusBar) SuggestPlayerFace(newState faceState) {
	if status.faceState == newState || status.faceState.priority > newState.priority {
		return
	}
	status.forcePlayerFace(newState)
}

func (status *statusBar) init() {
	*status = statusBar{}

	// Left HUD panel
	leftPanelTex := cache.GetTexture("assets/textures/ui/hud_backdrop_left.png")
	panelHeight := float32(leftPanelTex.Height()) * settings.SpriteScale()
	status.leftPanel = ui.NewBoxFull(
		math2.Rect{
			X: 0.0, Y: settings.UIHeight() - panelHeight,
			Width:  float32(leftPanelTex.Width()) * settings.SpriteScale(),
			Height: panelHeight,
		},
		leftPanelTex,
		color.White,
		5.0,
	)

	fitToSlice := func(parent math2.Rect, slice textures.Slice) math2.Rect {
		return math2.Rect{
			X:      parent.X + slice.Bounds.X*settings.SpriteScale(),
			Y:      parent.Y + slice.Bounds.Y*settings.SpriteScale(),
			Width:  slice.Bounds.Width * settings.SpriteScale(),
			Height: slice.Bounds.Height * settings.SpriteScale(),
		}
	}

	hudIconsTexture := cache.GetTexture(texHudIcons)

	// Heart icon
	heartSlice := leftPanelTex.FindSlice("healthIcon")
	status.heartIcon = ui.Box{
		Color:   color.White,
		Texture: hudIconsTexture,
		Transform: ui.Transform{
			Dest:  fitToSlice(status.leftPanel.Dest, heartSlice),
			Depth: 6.0,
		},
	}
	if heartAnim, ok := hudIconsTexture.GetAnimation("heart"); ok {
		status.heartIcon.AnimPlayer.ChangeAnimation(heartAnim)
		status.heartIcon.AnimPlayer.PlayFromStart()
	}

	// Face
	status.faceState = FaceStateIdle
	faceTex := cache.GetTexture(texSeganFace)
	faceSlice := leftPanelTex.FindSlice("face")
	status.face = ui.Box{
		Color:   color.White,
		Texture: faceTex,
		Transform: ui.Transform{
			Dest:  fitToSlice(status.leftPanel.Dest, faceSlice),
			Depth: 6.0,
		},
	}
	if faceAnim, ok := faceTex.GetAnimation(FaceStateIdle.anim); ok {
		status.face.AnimPlayer.PlayNewAnim(faceAnim)
	}

	// Health counter
	counterFont, err := cache.GetFont("assets/textures/ui/hud_counter_font.fnt")
	if err != nil {
		log.Println("Could not load counter font for HUD.", err)
	}
	healthStatSlice := leftPanelTex.FindSlice("healthStat")
	status.healthStat = ui.Text{
		Settings: ui.TextSettings{
			Font:      counterFont,
			Text:      "000",
			Alignment: ui.TEXT_ALIGN_CENTER,
		},
		Transform: ui.Transform{
			Dest:  fitToSlice(status.leftPanel.Dest, healthStatSlice),
			Depth: 6.0,
			Scale: settings.SpriteScale(),
		},
		Color: color.Red,
	}

	// Right HUD panel
	rightPanelTex := cache.GetTexture("assets/textures/ui/hud_backdrop_right.png")
	rightPanelWidth := rightPanelTex.Rect().Width * settings.SpriteScale()
	status.rightPanel = ui.NewBoxFull(
		math2.Rect{
			X:      settings.UIWidth() - rightPanelWidth,
			Y:      settings.UIHeight() - panelHeight,
			Width:  rightPanelWidth,
			Height: panelHeight,
		},
		rightPanelTex,
		color.White,
		5.0,
	)

	// Ammo icon
	ammoIconSlice := rightPanelTex.FindSlice("ammoIcon")
	status.ammoIcon = ui.Box{
		Color:   color.White,
		Texture: hudIconsTexture,
		Transform: ui.Transform{
			Dest:  fitToSlice(status.rightPanel.Dest, ammoIconSlice),
			Depth: 6.0,
		},
	}

	// Ammo counter
	ammoStatSlice := rightPanelTex.FindSlice("ammoStat")
	status.ammoStat = ui.Text{
		Settings: ui.TextSettings{
			Font:      counterFont,
			Text:      "000",
			Alignment: ui.TEXT_ALIGN_CENTER,
		},
		Transform: ui.Transform{
			Dest:  fitToSlice(status.rightPanel.Dest, ammoStatSlice),
			Depth: 6.0,
			Scale: settings.SpriteScale(),
		},
		Color: color.Blue,
	}

	// Armor icon
	armorIconSlice := rightPanelTex.FindSlice("armorIcon")
	status.armorIcon = ui.Box{
		Color:   color.White,
		Texture: hudIconsTexture,
		Transform: ui.Transform{
			Dest:  fitToSlice(status.rightPanel.Dest, armorIconSlice),
			Depth: 6.0,
		},
	}

	// Ammo counter
	armorStatSlice := rightPanelTex.FindSlice("armorStat")
	status.armorStat = ui.Text{
		Settings: ui.TextSettings{
			Font:      counterFont,
			Text:      "000",
			Alignment: ui.TEXT_ALIGN_CENTER,
		},
		Transform: ui.Transform{
			Dest:  fitToSlice(status.rightPanel.Dest, armorStatSlice),
			Depth: 6.0,
			Scale: settings.SpriteScale(),
		},
		Color: color.Green,
	}

	// Key icons
	for i, key := range [...]game.Keys{game.KeysBlue, game.KeysBrown, game.KeysYellow, game.KeysGray} {
		keyName := key.Name() + "Key"
		slice := rightPanelTex.FindSlice(keyName)
		status.keyIcons[i] = ui.Box{
			Color:   color.White,
			Texture: cache.GetTexture("assets/textures/ui/hud_keycards.png"),
			Transform: ui.Transform{
				Dest:  fitToSlice(status.rightPanel.Dest, slice),
				Depth: 6.0,
			},
		}
		switch key {
		case game.KeysBlue:
			status.keyIcons[i].Src = math2.Rect{X: 0, Y: 0, Width: 8, Height: 8}
		case game.KeysBrown:
			status.keyIcons[i].Src = math2.Rect{X: 8, Y: 0, Width: 8, Height: 8}
		case game.KeysYellow:
			status.keyIcons[i].Src = math2.Rect{X: 0, Y: 8, Width: 8, Height: 8}
		case game.KeysGray:
			status.keyIcons[i].Src = math2.Rect{X: 8, Y: 8, Width: 8, Height: 8}
		}
	}
}

func (status *statusBar) forcePlayerFace(newState faceState) {
	if status.faceState != newState {
		faceTex := cache.GetTexture(texSeganFace)
		anim, _ := faceTex.GetAnimation(newState.anim)
		status.face.AnimPlayer.PlayNewAnim(anim)
		status.face.FlippedHorz = newState.flipX
	}
	status.faceState = newState
	status.faceTimer = newState.showTime
}

func (status *statusBar) Layout(queue *ui.RenderQueue, deltaTime float32, stats PlayerStats, selectedWeapon *Weapon) {
	queue.Add(&status.leftPanel)
	queue.Add(&status.rightPanel)

	// Health stat
	status.healthStat.SetText(fmt.Sprintf("%03d", stats.Health))
	queue.Add(&status.healthStat)

	// Heart icon
	status.heartIcon.Update(deltaTime)
	queue.Add(&status.heartIcon)

	iconsTex := cache.GetTexture(texHudIcons)
	if selectedWeapon != nil {
		// Ammo stat
		status.ammoStat.SetText(fmt.Sprintf("%03d", stats.Ammo[selectedWeapon.ammoType]))
		queue.Add(&status.ammoStat)

		// Ammo icon
		status.ammoIcon.Update(deltaTime)
		anim, ok := iconsTex.GetAnimation(ammoTypeIconNames[selectedWeapon.ammoType])
		if ok {
			if !status.ammoIcon.AnimPlayer.IsPlayingAnim(anim) {
				status.ammoIcon.AnimPlayer.PlayNewAnim(anim)
			}
			queue.Add(&status.ammoIcon)
		}
	}

	// Armor stat
	if stats.Armor != game.ArmorTypeNone {
		status.armorStat.SetText(fmt.Sprintf("%03d", stats.ArmorAmount))
		queue.Add(&status.armorStat)
	}

	// Armor icon
	status.armorIcon.Update(deltaTime)
	anim, ok := iconsTex.GetAnimation(stats.Armor.Name() + "Armor")
	if ok {
		if !status.armorIcon.AnimPlayer.IsPlayingAnim(anim) {
			status.armorIcon.AnimPlayer.PlayNewAnim(anim)
		}
		queue.Add(&status.armorIcon)
	}

	// Keycards
	for i := range status.keyIcons {
		if (1<<i)&int(stats.Keys) != 0 {
			status.keyIcons[i].Update(deltaTime)
			queue.Add(&status.keyIcons[i])
		}
	}

	// Decide which face to display
	if stats.Health <= 0 {
		status.forcePlayerFace(FaceStateDead)
		status.heartIcon.AnimPlayer.Stop()
	} else if stats.GodMode {
		status.forcePlayerFace(FaceStateGod)
	} else if stats.Noclip {
		status.forcePlayerFace(FaceStateNoclip)
	} else {
		status.faceTimer -= deltaTime
		if status.faceTimer <= 0.0 {
			// Revert to idle face as a default
			status.forcePlayerFace(FaceStateIdle)
		}
	}
	status.face.Update(deltaTime)
	queue.Add(&status.face)
}
