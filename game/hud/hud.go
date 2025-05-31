package hud

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	MESSAGE_FADE_SPEED   = 2.0
	MESSAGE_SCROLL_SPEED = 0.1
	TEXT_FLICKER_SPEED   = 0.5
	VICTORY_COUNT_SPEED  = 0.1
	DEFAULT_FONT_PATH    = "assets/textures/ui/font.fnt"
	COUNTER_FONT_PATH    = "assets/textures/ui/hud_counter_font.fnt"
	SFX_STATS_DING       = "assets/sounds/ui/stats_ding.wav"
	LEVEL_INTRO_TIME     = 3.0 // Time after which the level intro ends.
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
	parusu                     Parusu
	airhorn                    Airhorn
	weapons                    [WEAPON_ORDER_COUNT]Weapon
	selectedWeapon, nextWeapon WeaponIndex

	flickerTime float32
	intro       struct {
		Timer                            float32
		Sweep1, Sweep2, Banner1, Banner2 scene.Id[*ui.Box]
		Background, Star, Sickle, Eyes   scene.Id[*ui.Box]
		Title, MapNumber                 scene.Id[*ui.Text]
		Voice                            tdaudio.VoiceId
	}
}

func (hud *Hud) Init() {
	hud.UI = ui.NewUIScene(256, 64)

	var fpsText *ui.Text
	hud.FPSCounter, fpsText, _ = hud.UI.Texts.New()
	fpsText.Dest = math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0}

	var spriteCounter *ui.Text
	hud.SpriteCounter, spriteCounter, _ = hud.UI.Texts.New()
	spriteCounter.Dest = math2.Rect{X: 4.0, Y: 56.0, Width: 480.0, Height: 128.0}
	spriteCounter.Color = color.Blue

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

	hud.InitPlayerStats()
}

// Sets up the UI elements on the victory screen
func (hud *Hud) InitVictory() {
	hud.UI.Boxes.Clear()
	hud.UI.Texts.Clear()

	_, completeTxt, _ := hud.UI.Texts.New()
	completeTxt.Transform = ui.Transform{
		Dest: math2.Rect{
			X:      settings.UIWidth()/4.0 - 32.0,
			Y:      24.0,
			Width:  settings.UIWidth() / 2.0,
			Height: 64.0,
		},
		Scale: 3.0,
	}
	completeTxt.Settings = ui.TextSettings{
		Text:         settings.Localize("levelComplete"),
		ShadowColor:  settings.Current.TextShadowColor,
		ShadowOffset: mgl32.Vec2{2.0, 2.0},
		Font:         cache.DefaultFont,
	}

	hud.KillsCounted = 0
	hud.SecretsCounted = 0

	var levelStats *ui.Text
	hud.levelStats, levelStats, _ = hud.UI.Texts.New()
	levelStats.Transform = ui.Transform{
		Dest: math2.Rect{
			X:      64.0,
			Y:      108.0,
			Width:  256.0,
			Height: 256.0,
		},
		Scale: 2.0,
	}
	levelStats.SetShadow(settings.Current.TextShadowColor, mgl32.Vec2{2.0, 2.0})
}

func (hud *Hud) InitIntro(levelTitle, mapNumber string) {
	if len(levelTitle) == 0 || len(mapNumber) == 0 {
		hud.intro.Timer = math2.Inf32()
		return
	}

	hud.intro.Background, _, _ = hud.UI.Boxes.New(ui.Box{
		Color: color.Black,
		Transform: ui.Transform{
			Dest:  math2.Rect{Width: settings.UIWidth(), Height: settings.UIHeight()},
			Depth: 8.9,
		},
	})

	starMesh, _ := cache.GetMesh("assets/models/star_transition.obj")
	hud.intro.Star, _, _ = hud.UI.Boxes.New(ui.Box{
		Transform: ui.Transform{
			Dest: math2.Rect{
				Y:      -settings.UIHeight() * 0.25,
				Width:  settings.UIWidth(),
				Height: settings.UIHeight() * 1.5,
			},
			Depth: 8.9,
			Scale: 0.5,
		},
		Color:  color.Black,
		Mesh:   starMesh,
		Hidden: true,
	})

	hud.intro.Banner1, _, _ = hud.UI.Boxes.New(ui.Box{
		Color: color.Blue,
		Transform: ui.Transform{
			Dest:  math2.Rect{X: 0.0, Y: 64.0, Width: settings.UIWidth(), Height: 96.0},
			Depth: 9.0,
		},
	})

	var titleTxt *ui.Text
	hud.intro.Title, titleTxt, _ = hud.UI.Texts.New()
	titleTxt.Settings = ui.TextSettings{
		Text:      levelTitle,
		Alignment: ui.TEXT_ALIGN_CENTER,
		Font:      cache.DefaultFont,
	}
	titleTxt.Transform = ui.Transform{
		Dest: math2.Rect{
			Y:      80.0,
			Width:  settings.UIWidth(),
			Height: 64.0,
		},
		Scale: 3.0,
		Depth: 9.1,
	}

	var sweep1 *ui.Box
	hud.intro.Sweep1, sweep1, _ = hud.UI.Boxes.New()
	sweep1.Transform = ui.Transform{
		Dest: math2.Rect{
			X:      0.0,
			Y:      63.0,
			Width:  settings.UIWidth(),
			Height: 98.0,
		},
		Depth: 9.2,
	}
	sweep1.Color = color.Black

	var banner2 *ui.Box
	hud.intro.Banner2, banner2, _ = hud.UI.Boxes.New(ui.Box{
		Color: color.Blue,
		Transform: ui.Transform{
			Dest:  math2.Rect{X: -48.0, Y: settings.UIHeight() - 224.0, Width: 352.0, Height: 96.0},
			Depth: 9.0,
			Shear: 1.0,
		},
	})

	var episodeTxt *ui.Text
	hud.intro.MapNumber, episodeTxt, _ = hud.UI.Texts.New()
	episodeTxt.SetText(mapNumber)
	episodeTxt.Transform = ui.Transform{
		Dest:  math2.Rect{X: 32.0, Y: banner2.DestPosition()[1] + 8.0, Width: 256.0, Height: banner2.Dest.Height},
		Scale: 3.0,
		Depth: 9.1,
	}

	var sweep2 *ui.Box
	hud.intro.Sweep2, sweep2, _ = hud.UI.Boxes.New()
	sweep2.Transform = ui.Transform{
		Dest: math2.Rect{
			X:      0.0,
			Y:      banner2.Dest.Y - 1.0,
			Width:  banner2.Dest.Width + 1.0,
			Height: banner2.Dest.Height + 2.0,
		},
		Depth: 9.2,
	}
	sweep2.Color = color.Black

	sickleTex := cache.GetTexture("assets/textures/ui/intro_sickle.png")
	var sickle *ui.Box
	hud.intro.Sickle, sickle, _ = hud.UI.Boxes.New(ui.NewBoxFull(math2.Rect{
		X:      settings.UIWidth() + 8.0,
		Y:      sweep1.Dest.Y - float32(sickleTex.Height())/2.0,
		Width:  float32(sickleTex.Width()) * SpriteScale(),
		Height: float32(sickleTex.Height()) * SpriteScale(),
	}, sickleTex, color.White))
	sickle.Depth = 9.3

	var eyes *ui.Box
	hud.intro.Eyes, eyes, _ = hud.UI.Boxes.New()
	eyesTex := cache.GetTexture("assets/textures/ui/intro_eyes.png")
	eyes.AnimPlayer = comps.NewAnimationPlayer(eyesTex.GetDefaultAnimation(), true)
	eyes.Texture = eyesTex
	eyesWidth := eyes.AnimPlayer.Frame().Rect.Width * SpriteScale()
	eyesHeight := eyes.AnimPlayer.Frame().Rect.Height * SpriteScale()
	eyes.Transform = ui.Transform{
		Dest: math2.Rect{
			X:      settings.UIWidth()/2.0 - eyesWidth/2.0,
			Y:      settings.UIHeight()/2.0 - eyesHeight/2.0,
			Width:  eyesWidth,
			Height: eyesHeight,
		},
		Depth: 9.3,
	}
}

func (hud *Hud) Update(deltaTime float32) {
	hud.UI.Update(deltaTime)

	hud.flickerTime += deltaTime
	if continueTxt, ok := hud.continueText.Get(); ok {
		continueTxt.Hidden = math2.Mod(hud.flickerTime, TEXT_FLICKER_SPEED) > TEXT_FLICKER_SPEED*0.75
	}

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
	if hud.intro.Timer < LEVEL_INTRO_TIME {
		hud.intro.Timer += deltaTime

		sickle, _ := scene.Get[*ui.Box](hud.intro.Sickle.Handle)
		sickleSpeed := deltaTime * settings.UIWidth() * 3.0
		sickle.Rotation += deltaTime * math.Pi * 16.0
		switch {
		case hud.intro.Timer < 0.5:
			// Wait
			if !hud.intro.Voice.IsValid() {
				hud.intro.Voice = cache.GetSfx("assets/sounds/ui/intro_whoosh1.wav").Play()
			}
		case hud.intro.Timer < 1.0:
			hud.intro.Voice = tdaudio.VoiceId{}
			if sickle.Dest.X < -float32(sickle.Texture.Width())*SpriteScale() {
				sweep2, _ := scene.Get[*ui.Box](hud.intro.Sweep2.Handle)
				sickle.Dest.Y = sweep2.Dest.Y - float32(sickle.Texture.Height()/2)*SpriteScale()
			} else {
				sickle.Dest.X -= sickleSpeed
			}
		case hud.intro.Timer < 3.5:
			if !hud.intro.Voice.IsValid() {
				hud.intro.Voice = cache.GetSfx("assets/sounds/ui/intro_whoosh2.wav").Play()
			}
			sickle.Dest.X += sickleSpeed
		}

		switch {
		case hud.intro.Timer < 0.5:
			// Wait
		case hud.intro.Timer < 1.0:
			sweep1, _ := scene.Get[*ui.Box](hud.intro.Sweep1.Handle)
			delta := deltaTime * settings.UIWidth() * 3.0
			sweep1.Dest.Width -= delta
		case hud.intro.Timer < 2.0:
			sweep2, _ := scene.Get[*ui.Box](hud.intro.Sweep2.Handle)
			delta := deltaTime * 2048.0
			sweep2.Dest.Width = max(0, sweep2.Dest.Width-delta)
			sweep2.Dest.X += delta
		case hud.intro.Timer < 2.5:
			bg, _ := scene.Get[*ui.Box](hud.intro.Background.Handle)
			bg.Hidden = true
			star, _ := scene.Get[*ui.Box](hud.intro.Star.Handle)
			star.Hidden = false
			star.Scale += deltaTime * 200.0
		case hud.intro.Timer < 3.5:
			titleBh, _ := scene.Get[*ui.Box](hud.intro.Banner1.Handle)
			delta := deltaTime * settings.UIWidth() * 2.0
			titleBh.Dest.X += delta
			titleTxt, _ := scene.Get[*ui.Text](hud.intro.Title.Handle)
			titleTxt.Dest.X += delta
			episodeBanner, _ := scene.Get[*ui.Box](hud.intro.Banner2.Handle)
			episodeBanner.Dest.X -= delta
			episodeTxt, _ := scene.Get[*ui.Text](hud.intro.MapNumber.Handle)
			episodeTxt.Dest.X -= delta
		case hud.intro.Timer >= LEVEL_INTRO_TIME:
			// Remove all elements in the intro struct by calling the Remove method.
			introStruct := reflect.ValueOf(hud.intro)
			for f := range introStruct.NumField() {
				field := introStruct.Field(f)
				if method := field.MethodByName("Remove"); method != (reflect.Value{}) {
					method.Call(nil)
				}
			}
		}
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
						txt.Transform = ui.Transform{
							Dest:  math2.RectFromRadius(settings.UIWidth()/2.0, 7.0*settings.UIHeight()/8.0, 256.0, 48.0),
							Scale: 2.0,
						}
						txt.Settings = ui.TextSettings{
							Text:         settings.Localize("fireContinue"),
							Alignment:    ui.TEXT_ALIGN_CENTER,
							ShadowColor:  settings.Current.TextShadowColor,
							ShadowOffset: mgl32.Vec2{2.0, 2.0},
							Font:         cache.DefaultFont,
						}
						txt.Color = color.Color{R: 0.9, G: 0.9, B: 0, A: 1.0}
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

func (hud *Hud) IntroTimeLeft() float32 {
	return LEVEL_INTRO_TIME - hud.intro.Timer
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
