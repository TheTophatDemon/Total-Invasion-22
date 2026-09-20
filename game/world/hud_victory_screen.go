package world

import (
	"fmt"
	"strings"
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
	timeCounter, killCounter, secretCounter counter
	currentCounter                          *counter // Refers to which of the above counters are being counted
	countTimer                              float32  // Seconds between counting the stats on the victory screen
	flickerTime                             float32  // Timer for flickering text

	txtComplete, txtContinue, txtStats ui.Element
}

func (screen *VictoryScreen) init() {
	*screen = VictoryScreen{
		currentCounter: &screen.timeCounter,
	}

	chungusFont, err := cache.GetFont("assets/textures/ui/chungus_font.fnt")
	if err != nil {
		chungusFont = cache.DefaultFont
	}

	// Level complete text
	screen.txtComplete = ui.NewText(ui.Transform{
		Origin:   ui.Ratios{0.5, 0.0},
		Anchor:   ui.Ratios{0.5, 0.0},
		Position: mgl32.Vec2{0.0, 12.0},
		Size:     mgl32.Vec2{settings.UIWidth(), 96.0},
	}, settings.Localize("levelComplete"), ui.DefaultTextConfig().
		SetAlign(ui.TextAlignCenterH|ui.TextAlignCenterV).
		SetScale(2.0).SetFont(chungusFont),
	)

	// Continue prompt
	screen.txtContinue = ui.NewText(ui.Transform{
		Origin:   ui.Ratios{0.5, 0.5},
		Anchor:   ui.Ratios{0.5, 1.0},
		Position: mgl32.Vec2{0.0, -64.0},
		Size:     mgl32.Vec2{512.0, 48.0},
	}, settings.Localize("fireContinue"), ui.DefaultTextConfig().
		SetColor(color.Color{R: 0.9, G: 0.9, B: 0, A: 1.0}).
		SetAlign(ui.TextAlignCenterH|ui.TextAlignCenterV).
		SetWrapWords(true),
	)

	// Stats text
	screen.txtStats = ui.NewText(ui.Transform{
		Position: mgl32.Vec2{64.0, 108.0},
		Size:     mgl32.Vec2{512.0, 512.0},
	}, "(filled in later)", ui.DefaultTextConfig().SetScale(2.0))
}

func (screen *VictoryScreen) EndLevel() {
	screen.timeCounter = counter{
		max:  gWorld.levelEndTime.Sub(gWorld.levelStartTime).Milliseconds(),
		step: 12_000,
	}
	screen.killCounter = counter{
		max:  int64(gWorld.enemiesKilled),
		step: 1,
	}
	screen.secretCounter = counter{
		max:  int64(gWorld.secretsFound),
		step: 1,
	}
}

func (screen *VictoryScreen) Layout(queue *ui.RenderQueue, deltaTime float32) {

	queue.Add(&screen.txtComplete)

	const TEXT_FLICKER_SPEED = 0.5
	screen.flickerTime += deltaTime
	if screen.currentCounter == nil && math2.Mod(screen.flickerTime/TEXT_FLICKER_SPEED, 1.0) < 0.75 {
		queue.Add(&screen.txtContinue)
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
	var statsText strings.Builder
	statsText.Grow(64)
	statsText.WriteString(settings.Localize("statTime"))
	fmt.Fprintf(&statsText, ": %02d:%05.2f\n", int(timeStat.Minutes()), math2.Mod(timeStat.Seconds(), 60.0))
	statsText.WriteString(settings.Localize("statKills"))
	fmt.Fprintf(&statsText, ": %02d/%02d\n", screen.killCounter.count, gWorld.enemiesTotal)
	statsText.WriteString(settings.Localize("statSecrets"))
	fmt.Fprintf(&statsText, ": %02d/%02d", screen.secretCounter.count, gWorld.secretsTotal)
	screen.txtStats.SetText(statsText.String())
	queue.Add(&screen.txtStats)
}
