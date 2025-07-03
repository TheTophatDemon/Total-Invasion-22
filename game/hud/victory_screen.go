package hud

import (
	"fmt"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type counter struct {
	count, max, step int64
}

type VictoryScreen struct {
	EnemiesKilled, EnemiesTotal uint
	SecretsFound, SecretsTotal  uint

	levelStartTime, levelEndTime            time.Time
	timeCounter, killCounter, secretCounter counter
	currentCounter                          *counter // Refers to which of the above counters are being counted
	countTimer                              float32  // Seconds between counting the stats on the victory screen
	flickerTime                             float32  // Timer for flickering text
}

func (screen *VictoryScreen) init() {
	*screen = VictoryScreen{
		levelStartTime: time.Now(),
		currentCounter: &screen.timeCounter,
	}
}

func (screen *VictoryScreen) EndLevel() {
	screen.levelEndTime = time.Now()
	screen.timeCounter = counter{
		max:  screen.levelEndTime.Sub(screen.levelStartTime).Milliseconds(),
		step: 12_000,
	}
	screen.killCounter = counter{
		max:  int64(screen.EnemiesKilled),
		step: 1,
	}
	screen.secretCounter = counter{
		max:  int64(screen.SecretsFound),
		step: 1,
	}
}

func (screen *VictoryScreen) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if screen.levelEndTime.IsZero() {
		// Only show after level ends.
		return
	}

	// Level complete text
	queue.Add(&ui.Text{
		Color: color.White,
		Settings: ui.TextSettings{
			Text:         settings.Localize("levelComplete"),
			ShadowColor:  settings.Current.TextShadowColor,
			ShadowOffset: mgl32.Vec2{2.0, 2.0},
			Font:         cache.DefaultFont,
		},
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      settings.UIWidth()/4.0 - 32.0,
				Y:      24.0,
				Width:  settings.UIWidth() / 2.0,
				Height: 64.0,
			},
			Scale: 3.0,
		},
	})

	const TEXT_FLICKER_SPEED = 0.5
	screen.flickerTime += deltaTime
	// Continue prompt
	if screen.currentCounter == nil && math2.Mod(screen.flickerTime/TEXT_FLICKER_SPEED, 1.0) < 0.75 {
		queue.Add(&ui.Text{
			Color: color.Color{R: 0.9, G: 0.9, B: 0, A: 1.0},
			Settings: ui.TextSettings{
				Text:         settings.Localize("fireContinue"),
				Alignment:    ui.TEXT_ALIGN_CENTER,
				ShadowColor:  settings.Current.TextShadowColor,
				ShadowOffset: mgl32.Vec2{2.0, 2.0},
				Font:         cache.DefaultFont,
				WrapWords:    true,
			},
			Transform: ui.Transform{
				Dest:  math2.RectFromRadius(settings.UIWidth()/2.0, 7.0*settings.UIHeight()/8.0, 256.0, 48.0),
				Scale: 2.0,
			},
		})
	}

	screen.countTimer += deltaTime
	if screen.countTimer > 0.1 {
		screen.countTimer = 0.0
		if screen.currentCounter != nil {
			if screen.currentCounter.max > 0 {
				cache.GetSfx("assets/sounds/ui/stats_ding.wav").Play()
			}
			screen.currentCounter.count += screen.currentCounter.step
			if screen.currentCounter.count >= screen.currentCounter.max {
				screen.currentCounter.count = screen.currentCounter.max
				switch screen.currentCounter {
				case &screen.timeCounter:
					screen.currentCounter = &screen.killCounter
					screen.countTimer = -0.5
				case &screen.killCounter:
					screen.currentCounter = &screen.secretCounter
					screen.countTimer = -0.5
				case &screen.secretCounter:
					screen.currentCounter = nil
				}
			}
		}
	}

	// Level stats text
	timeStat := time.Duration(screen.timeCounter.count) * time.Millisecond
	queue.Add(&ui.Text{
		Color: color.White,
		Settings: ui.TextSettings{
			ShadowColor:  settings.Current.TextShadowColor,
			ShadowOffset: mgl32.Vec2{2.0, 2.0},
			Text: settings.Localize("statTime") + fmt.Sprintf(": %02d:%05.2f\n", int(timeStat.Minutes()), math2.Mod(timeStat.Seconds(), 60.0)) +
				settings.Localize("statKills") + fmt.Sprintf(": %02d/%02d\n", screen.killCounter.count, screen.EnemiesTotal) +
				settings.Localize("statSecrets") + fmt.Sprintf(": %02d/%02d", screen.secretCounter.count, screen.SecretsTotal),
		},
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      64.0,
				Y:      108.0,
				Width:  256.0,
				Height: 256.0,
			},
			Scale: 2.0,
		},
	})
}
