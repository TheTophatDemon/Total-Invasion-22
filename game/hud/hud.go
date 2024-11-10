package hud

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	MESSAGE_FADE_SPEED  = 2.0
	TEXT_FLICKER_SPEED  = 0.5
	VICTORY_COUNT_SPEED = 0.1
	DEFAULT_FONT_PATH   = "assets/textures/ui/font.fnt"
	COUNTER_FONT_PATH   = "assets/textures/ui/hud_counter_font.fnt"
	SFX_STATS_DING      = "assets/sounds/ui/stats_ding.wav"
)

type CountState uint8

const (
	COUNT_STATE_START CountState = iota
	COUNT_STATE_TIME
	COUNT_STATE_PAUSE1
	COUNT_STATE_KILLS
	COUNT_STATE_PAUSE2
	COUNT_STATE_SECRETS
	COUNT_STATE_DONE
)

type Hud struct {
	UI *ui.Scene

	LevelStartTime, LevelEndTime               time.Time
	LevelTimePercent                           float32
	KillsCounted, EnemiesKilled, EnemiesTotal  uint
	SecretsCounted, SecretsFound, SecretsTotal uint
	countTimer                                 float32 // Seconds between counting the stats on the victory screen
	countState                                 CountState

	FPSCounter, SpriteCounter scene.Id[*ui.Text]
	face                      scene.Id[*ui.Box]
	faceState                 faceState
	faceTimer                 float32
	heartIcon, ammoIcon       scene.Id[*ui.Box]
	keyIcons                  [4]scene.Id[*ui.Box]
	healthStat, ammoStat      scene.Id[*ui.Text]
	levelStats, continueText  scene.Id[*ui.Text]

	messageText     scene.Id[*ui.Text]
	messageTimer    float32
	messagePriority int

	flashRect  scene.Id[*ui.Box]
	flashSpeed float32

	sickle                     Sickle
	chickenGun                 ChickenCannon
	grenadeLauncher            GrenadeLauncher
	weapons                    [WEAPON_ORDER_COUNT]Weapon
	selectedWeapon, nextWeapon WeaponIndex

	flickerTime float32
}

func (hud *Hud) Init() {
	hud.UI = ui.NewUIScene(256, 64)

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
		SetDest(math2.Rect{
			X:      settings.UIWidth() / 4.0,
			Y:      settings.UIHeight() / 2.0,
			Width:  settings.UIWidth() / 2.0,
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

	hud.sickle.Init(hud)
	hud.chickenGun.Init(hud)
	hud.grenadeLauncher.Init(hud)
	hud.weapons = [WEAPON_ORDER_COUNT]Weapon{
		WEAPON_ORDER_SICKLE:  &hud.sickle,
		WEAPON_ORDER_CHICKEN: &hud.chickenGun,
		WEAPON_ORDER_GRENADE: &hud.grenadeLauncher,
	}

	hud.InitPlayerStats()
}

// Sets up the UI elements on the victory screen
func (hud *Hud) InitVictory() {
	hud.UI.Boxes.Clear()
	hud.UI.Texts.Clear()

	_, completeTxt, _ := hud.UI.Texts.New()
	completeTxt.
		SetFont(DEFAULT_FONT_PATH).
		SetText(settings.Localize("levelComplete")).
		SetDest(math2.Rect{
			X:      settings.UIWidth()/4.0 - 32.0,
			Y:      96.0,
			Width:  settings.UIWidth() / 2.0,
			Height: 64.0,
		}).
		SetScale(3.0).
		SetShadow(settings.Current.TextShadowColor, mgl32.Vec2{2.0, 2.0})

	hud.KillsCounted = 0
	hud.SecretsCounted = 0

	var levelStats *ui.Text
	hud.levelStats, levelStats, _ = hud.UI.Texts.New()
	levelStats.
		SetFont(DEFAULT_FONT_PATH).
		SetText("").
		SetDest(math2.Rect{
			X:      64.0,
			Y:      180.0,
			Width:  256.0,
			Height: 256.0,
		}).
		SetScale(2.0).
		SetShadow(settings.Current.TextShadowColor, mgl32.Vec2{2.0, 2.0})
}

func (hud *Hud) Update(deltaTime float32) {
	hud.UI.Update(deltaTime)

	hud.flickerTime += deltaTime
	if continueTxt, ok := hud.continueText.Get(); ok {
		continueTxt.Hidden = math2.Mod(hud.flickerTime, TEXT_FLICKER_SPEED) > TEXT_FLICKER_SPEED*0.75
	}

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

	// Update victory stats
	if levelStats, ok := hud.levelStats.Get(); ok {
		hud.countTimer += deltaTime
		if hud.countState != COUNT_STATE_DONE && (hud.countTimer > VICTORY_COUNT_SPEED || hud.countState == COUNT_STATE_START) {
			hud.countTimer = 0.0
			switch hud.countState {
			case COUNT_STATE_START:
				hud.countState++
			case COUNT_STATE_TIME:
				if hud.LevelTimePercent < 1.0 {
					cache.GetSfx(SFX_STATS_DING).Play()
					hud.LevelTimePercent += 0.1
				} else {
					hud.countState++
				}
			case COUNT_STATE_KILLS:
				if hud.KillsCounted < hud.EnemiesKilled {
					hud.KillsCounted++
					cache.GetSfx(SFX_STATS_DING).Play()
				} else {
					hud.countState++
				}
			case COUNT_STATE_SECRETS:
				if hud.SecretsCounted < hud.SecretsFound {
					cache.GetSfx(SFX_STATS_DING).Play()
					hud.SecretsCounted += 1
				} else {
					hud.countState++

					// Show continue prompt
					var txt *ui.Text
					var err error
					if hud.continueText, txt, err = hud.UI.Texts.New(); err == nil {
						txt.SetFont(DEFAULT_FONT_PATH).
							SetText(settings.Localize("fireContinue")).
							SetDest(math2.RectFromRadius(settings.UIWidth()/2.0, 7.0*settings.UIHeight()/8.0, 256.0, 24.0)).
							SetAlignment(ui.TEXT_ALIGN_CENTER).
							SetScale(2.0).
							SetColor(color.Color{R: 0.9, G: 0.9, B: 0, A: 1.0}).
							SetShadow(settings.Current.TextShadowColor, mgl32.Vec2{2.0, 2.0})
					}
				}
			default:
				hud.countState++
			}

			runTime := hud.LevelEndTime.Sub(hud.LevelStartTime)
			countedTime := hud.LevelStartTime.Add(time.Duration(float64(runTime.Nanoseconds()) * float64(hud.LevelTimePercent))).Sub(hud.LevelStartTime)

			var statsText strings.Builder
			statsText.Grow(256)
			statsText.WriteString(settings.Localize("statTime") + fmt.Sprintf(": %02d:%05.2f\n", int(countedTime.Minutes()), math2.Mod(countedTime.Seconds(), 60.0)))
			statsText.WriteString(settings.Localize("statKills") + fmt.Sprintf(": %02d/%02d\n", hud.KillsCounted, hud.EnemiesTotal))
			statsText.WriteString(settings.Localize("statSecrets") + fmt.Sprintf(": %02d/%02d\n", hud.SecretsCounted, hud.SecretsTotal))
			levelStats.SetText(statsText.String())
		}
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
