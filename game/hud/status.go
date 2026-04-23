package hud

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
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
	leftPanel, rightPanel, face     ui.Element
	faceState                       faceState
	faceTimer                       float32
	heartIcon, ammoIcon, armorIcon  ui.Element
	keyIcons                        [4]ui.Element
	healthStat, ammoStat, armorStat ui.Element
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
	panelHeight := float32(leftPanelTex.Height()) * SpriteScale()
	status.leftPanel = ui.NewBox(
		ui.Transform{
			Position: mgl32.Vec2{0, -32.0},
			Anchor:   ui.Ratios{0.0, 1.0},
			Origin:   ui.Ratios{0.0, 1.0},
			Size:     mgl32.Vec2{float32(leftPanelTex.Width()) * SpriteScale(), panelHeight},
			Depth:    5.0,
		},
		leftPanelTex,
	)

	fitToSlice := func(parent *ui.Element, sliceName string, depth float32) ui.Transform {
		tex := parent.BgTexture
		slice := tex.FindSlice(sliceName)
		screenOfs := math2.ElemMul2(
			mgl32.Vec2{float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight)},
			mgl32.Vec2(parent.Anchor()),
		)
		originOfs := math2.ElemMul2(parent.Size(), mgl32.Vec2(parent.Origin()))
		parentTopLeft := parent.Position().Add(screenOfs).Sub(originOfs)
		return ui.Transform{
			Position: parentTopLeft.Add(slice.Bounds.PosVec().Mul(SpriteScale())),
			Size:     slice.Bounds.SizeVec().Mul(SpriteScale()),
			Depth:    depth,
		}
	}

	hudIconsTexture := cache.GetTexture(texHudIcons)

	// Heart icon
	status.heartIcon = ui.NewBox(fitToSlice(&status.leftPanel, "healthIcon", 6), hudIconsTexture)
	if heartAnim, ok := hudIconsTexture.GetAnimation("heart"); ok {
		status.heartIcon.AnimPlayer.ChangeAnimation(heartAnim)
		status.heartIcon.AnimPlayer.PlayFromStart()
	}

	// Face
	status.faceState = FaceStateIdle
	faceTex := cache.GetTexture(texSeganFace)
	status.face = ui.NewBox(fitToSlice(&status.leftPanel, "face", 6.0), faceTex)
	if faceAnim, ok := faceTex.GetAnimation(FaceStateIdle.anim); ok {
		status.face.AnimPlayer.PlayNewAnim(faceAnim)
	}

	// Health counter
	counterFont, _ := cache.GetFont("assets/textures/ui/hud_counter_font.fnt")
	status.healthStat = ui.NewText(
		fitToSlice(&status.leftPanel, "healthStat", 6.0),
		"000",
		ui.TextConfig{
			Font:          counterFont,
			Align:         ui.TextAlignCenterH | ui.TextAlignBottom,
			DisableShadow: true,
			Color:         maybe.Some(color.Red),
			Scale:         maybe.Some[float32](2.0),
		},
	)

	// Right HUD panel
	rightPanelTex := cache.GetTexture("assets/textures/ui/hud_backdrop_right.png")
	status.rightPanel = ui.NewBox(
		ui.Transform{
			Position: mgl32.Vec2{0.0, -32.0},
			Anchor:   ui.Ratios{1.0, 1.0},
			Origin:   ui.Ratios{1.0, 1.0},
			Size:     mgl32.Vec2{rightPanelTex.Rect().Width * SpriteScale(), panelHeight},
			Depth:    5.0,
		},
		rightPanelTex,
	)

	// Ammo icon
	status.ammoIcon = ui.NewBox(
		fitToSlice(&status.rightPanel, "ammoIcon", 6.0),
		hudIconsTexture,
	)

	// Ammo counter
	status.ammoStat = ui.NewText(
		fitToSlice(&status.rightPanel, "ammoStat", 6.0),
		"000",
		ui.TextConfig{
			Font:          counterFont,
			Align:         ui.TextAlignCenterH | ui.TextAlignBottom,
			DisableShadow: true,
			Color:         maybe.Some(color.Blue),
			Scale:         maybe.Some[float32](2.0),
		},
	)

	// Armor icon
	status.armorIcon = ui.NewBox(
		fitToSlice(&status.rightPanel, "armorIcon", 6.0),
		hudIconsTexture,
	)

	// Ammo counter
	status.armorStat = ui.NewText(
		fitToSlice(&status.rightPanel, "armorStat", 6.0),
		"000",
		ui.TextConfig{
			Font:          counterFont,
			Align:         ui.TextAlignCenterH | ui.TextAlignBottom,
			DisableShadow: true,
			Color:         maybe.Some(color.Green),
			Scale:         maybe.Some[float32](2.0),
		},
	)

	// Key icons
	keysTexture := cache.GetTexture("assets/textures/ui/hud_keycards.png")
	for i, key := range [...]game.Keys{game.KeysBlue, game.KeysBrown, game.KeysYellow, game.KeysGray} {
		status.keyIcons[i] = ui.NewBox(
			fitToSlice(&status.rightPanel, key.Name()+"Key", 6.0),
			keysTexture,
		)
		switch key {
		case game.KeysBlue:
			status.keyIcons[i].AnimPlayer.PlaySingleFrame(keysTexture, math2.Rect{X: 0, Y: 0, Width: 8, Height: 8})
		case game.KeysBrown:
			status.keyIcons[i].AnimPlayer.PlaySingleFrame(keysTexture, math2.Rect{X: 8, Y: 0, Width: 8, Height: 8})
		case game.KeysYellow:
			status.keyIcons[i].AnimPlayer.PlaySingleFrame(keysTexture, math2.Rect{X: 0, Y: 8, Width: 8, Height: 8})
		case game.KeysGray:
			status.keyIcons[i].AnimPlayer.PlaySingleFrame(keysTexture, math2.Rect{X: 8, Y: 8, Width: 8, Height: 8})
		}
	}
}

func (status *statusBar) forcePlayerFace(newState faceState) {
	if status.faceState != newState {
		faceTex := cache.GetTexture(texSeganFace)
		anim, _ := faceTex.GetAnimation(newState.anim)
		status.face.AnimPlayer.PlayNewAnim(anim)
		status.face.BgFlippedHorz = newState.flipX
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
	status.heartIcon.AnimPlayer.Update(deltaTime)
	queue.Add(&status.heartIcon)

	iconsTex := cache.GetTexture(texHudIcons)
	if selectedWeapon != nil {
		// Ammo stat
		status.ammoStat.SetText(fmt.Sprintf("%03d", stats.Ammo[selectedWeapon.ammoType]))
		queue.Add(&status.ammoStat)

		// Ammo icon
		status.ammoIcon.AnimPlayer.Update(deltaTime)
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
	status.armorIcon.AnimPlayer.Update(deltaTime)
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
			status.keyIcons[i].AnimPlayer.Update(deltaTime)
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
	status.face.AnimPlayer.Update(deltaTime)
	queue.Add(&status.face)
}
