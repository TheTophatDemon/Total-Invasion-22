package hud

import (
	"fmt"
	"unicode/utf8"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/hud/weapons"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	MESSAGE_FADE_SPEED   = 2.0
	MESSAGE_SCROLL_SPEED = 0.1
	DEFAULT_FONT_PATH    = "assets/textures/ui/font.fnt"
	COUNTER_FONT_PATH    = "assets/textures/ui/hud_counter_font.fnt"
	SFX_STATS_DING       = "assets/sounds/ui/stats_ding.wav"
)

type Hud struct {
	UI *ui.Scene

	renderQueue ui.RenderQueue
	weaponWheel WeaponWheel

	FPSCounter, SpriteCounter       scene.Id[*ui.Text]
	face                            scene.Id[*ui.Box]
	faceState                       faceState
	faceTimer                       float32
	heartIcon, ammoIcon, armorIcon  scene.Id[*ui.Box]
	keyIcons                        [4]scene.Id[*ui.Box]
	healthStat, ammoStat, armorStat scene.Id[*ui.Text]

	messageText     scene.Id[*ui.Text]
	messageTimer    float32
	messagePriority int

	flashRect  scene.Id[*ui.Box]
	flashSpeed float32

	weapons                    [weapons.WeaponCount]weapons.Weapon
	selectedWeapon, nextWeapon weapons.WeaponKind

	Intro         LevelIntro
	VictoryScreen VictoryScreen
}

func (hud *Hud) Init(debug bool) {
	hud.UI = ui.NewUIScene(256, 64)

	if debug {
		var fpsText *ui.Text
		hud.FPSCounter, fpsText, _ = hud.UI.Texts.New()
		fpsText.Dest = math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0}

		var spriteCounter *ui.Text
		hud.SpriteCounter, spriteCounter, _ = hud.UI.Texts.New()
		spriteCounter.Dest = math2.Rect{X: 4.0, Y: 56.0, Width: 480.0, Height: 128.0}
		spriteCounter.Color = color.Blue
	}

	leftPanelTex := cache.GetTexture("assets/textures/ui/hud_backdrop_left.png")
	rightPanelTex := cache.GetTexture("assets/textures/ui/hud_backdrop_right.png")

	_, messageBackground, _ := hud.UI.Boxes.New(ui.Box{
		Color: color.Black,
		Src: math2.Rect{
			Width: 1.0, Height: 1.0,
		},
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      float32(leftPanelTex.Width()) * SpriteScale(),
				Y:      settings.UIHeight() - 32.0,
				Width:  settings.UIWidth() - float32(leftPanelTex.Width()+rightPanelTex.Width())*SpriteScale(),
				Height: 32.0,
			},
			Depth: 2.0,
		},
	})

	var message *ui.Text
	hud.messageText, message, _ = hud.UI.Texts.New()
	message.Transform = ui.Transform{
		Dest: math2.Rect{
			X:      messageBackground.Dest.X + 8.0,
			Y:      messageBackground.Dest.Y + 1.0,
			Width:  messageBackground.Dest.Width - 16.0,
			Height: messageBackground.Dest.Height - 2.0,
		},
		Depth: 3.0,
		Scale: 1.0,
	}
	message.Settings.WrapWords = false

	hud.flashRect, _, _ = hud.UI.Boxes.New(ui.Box{
		Transform: ui.Transform{
			Dest: math2.Rect{
				Width:  settings.UIWidth(),
				Height: settings.UIHeight(),
			},
			Depth: 9.0,
			Scale: 1.0,
		},
		Color: color.Transparent,
	})

	hud.weapons = [WEAPON_ORDER_COUNT]Weapon{
		WEAPON_ORDER_SICKLE:  &hud.sickle,
		WEAPON_ORDER_CHICKEN: &hud.chickenGun,
		WEAPON_ORDER_GRENADE: &hud.grenadeLauncher,
		WEAPON_ORDER_PARUSU:  &hud.parusu,
		WEAPON_ORDER_AIRHORN: &hud.airhorn,
	}
	for _, weapon := range hud.weapons {
		if weapon != nil {
			weapon.Init(hud)
		}
	}

	hud.VictoryScreen.Init()
	hud.InitPlayerStats()
}

func (hud *Hud) Update(deltaTime float32) {
	hud.UI.Update(deltaTime)

	// Update message text
	if message, ok := hud.messageText.Get(); ok {
		hud.messageTimer -= deltaTime
		if hud.messageTimer <= 0.0 {
			if hud.messageTimer < -MESSAGE_SCROLL_SPEED {
				hud.messageTimer = 0.0
				msgText := message.Text()
				if len(msgText) > 1 {
					_, byteCount := utf8.DecodeRuneInString(msgText)
					message.SetText(msgText[byteCount:])
				} else {
					hud.messagePriority = 0
					message.Color = color.Transparent
					message.SetText("")
				}
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

	// Update level intro
	if !hud.Intro.Done() {
		hud.Intro.Layout(&hud.renderQueue, deltaTime)
	}

	// Update victory stats
	hud.VictoryScreen.Layout(&hud.renderQueue, deltaTime)
}

func (hud *Hud) UpdateDebugCounters(renderContext *render.Context, avgCollisionTime int64) {
	if sprCountTxt, ok := hud.SpriteCounter.Get(); ok {
		sprCountTxt.SetText(
			fmt.Sprintf("Sprites drawn: %v\nWalls drawn: %v\nParticles drawn: %v\nAvg. Collision MS: %v",
				renderContext.DrawnSpriteCount,
				renderContext.DrawnWallCount,
				renderContext.DrawnParticlesCount,
				avgCollisionTime))
	}
}

func (hud *Hud) Render() {
	// Setup 2D render context
	renderContext := render.Context{
		View:       mgl32.Ident4(),
		Projection: mgl32.Ortho(0.0, float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight), 0.0, -50.0, 50.0),
	}

	// Render 2D game elements
	hud.UI.Render(&renderContext)

	hud.renderQueue.Render(&renderContext)
}

func (hud *Hud) ShowMessage(text string, duration float32, priority int, colr color.Color) {
	if priority >= hud.messagePriority {
		hud.messageTimer = duration
		hud.messagePriority = priority
		if message, ok := hud.messageText.Get(); ok {
			message.SetText(text)
			message.Color = colr
		}
	}
}

func (hud *Hud) FlashScreen(color color.Color, fadeSpeed float32) {
	if flash, ok := hud.flashRect.Get(); ok {
		flash.Color = color
		hud.flashSpeed = fadeSpeed
	}
}

func (hud *Hud) Weapon(index WeaponIndex) Weapon {
	if index == WEAPON_ORDER_NONE {
		return nil
	}
	return hud.weapons[index]
}

func (hud *Hud) SelectedWeapon() Weapon {
	return hud.Weapon(hud.selectedWeapon)
}

func (hud *Hud) AttemptFireWeapon(ammo *game.Ammo) bool {
	weapon := hud.SelectedWeapon()
	if weapon != nil && weapon.CanFire(ammo) {
		weapon.Fire(ammo)
		return true
	}
	return false
}

func (hud *Hud) SelectWeapon(order WeaponIndex) {
	if order == hud.selectedWeapon || (order >= 0 && !hud.weapons[order].IsEquipped()) {
		return
	}
	if hud.selectedWeapon >= 0 {
		hud.weapons[hud.selectedWeapon].Deselect()
	}
	hud.nextWeapon = order
}

func (hud *Hud) EquipWeapon(order WeaponIndex) {
	if order < 0 || hud.weapons[order] == nil {
		return
	}
	hud.weapons[order].Equip()
}

func SpriteScale() float32 {
	return settings.UIScale() * 2.0
}
