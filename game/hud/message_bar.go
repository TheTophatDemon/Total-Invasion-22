package hud

import (
	"unicode/utf8"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tween"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type MessageBar struct {
	text     ui.Element
	priority int
	flash    tween.Data     // Tracks the color change in the message bar after a message is shown
	scroll   tween.Sequence // Tracks the appearance of the text as it scrolls off screen.
}

func (messageBar *MessageBar) init() {
	messageBar.text = ui.NewText(
		ui.Transform{
			Anchor: ui.Ratios{0.0, 1.0},
			Origin: ui.Ratios{0.0, 1.0},
			Size:   mgl32.Vec2{float32(settings.Current.WindowWidth), 32.0},
			Depth:  2.0,
		},
		"",
		ui.TextConfig{},
	)
	messageBar.text.BgMesh = cache.QuadMesh
	messageBar.text.BgColor = maybe.Some(color.Black)
}

func (messageBar *MessageBar) ShowMessage(text string, priority int, colr color.Color) {
	if priority >= messageBar.priority {
		charCount := utf8.RuneCount([]byte(text))
		tweens := make([]tween.Data, 0, charCount+1)
		// Pause to let the player read
		tweens = append(tweens, tween.Data{Duration: float32(charCount) * (3.0 / 80.0)})
		// Adds a tween for removing each character in sequence
		for i := range charCount {
			tweens = append(tweens, tween.Data{StartValue: float32(i), EndValue: float32(i + 1), Duration: 0.1})
		}
		messageBar.scroll = tween.Sequence{
			Tweens: tweens,
		}

		messageBar.priority = priority
		messageBar.text.SetText(text)
		messageBar.text.SetTextConfig(ui.TextConfig{
			Color:         maybe.Some(colr),
			WrapWords:     false,
			DisableShadow: true,
		})
		messageBar.flash = tween.Data{
			StartValue: 1.0,
			EndValue:   0.0,
			Duration:   0.5,
		}
	}
}

func (messageBar *MessageBar) layout(queue *ui.RenderQueue, deltaTime float32) {
	if messageBar.scroll.Update(deltaTime).TweenDone {
		msgText := messageBar.text.Text()
		if len(msgText) > 1 {
			_, byteCount := utf8.DecodeRuneInString(msgText)
			messageBar.text.SetText(msgText[byteCount:])
		} else {
			messageBar.priority = 0
			messageBar.text.SetText("")
		}
	}

	flashAmt := messageBar.flash.Update(deltaTime).Value
	txtColor := messageBar.text.TextConfig().Unwrap().Color.Or(color.White)
	messageBar.text.BgColor = maybe.Some(color.Color{
		R: (1.0 - txtColor.R) * flashAmt,
		G: (1.0 - txtColor.G) * flashAmt,
		B: (1.0 - txtColor.B) * flashAmt,
		A: 1.0,
	})

	queue.Add(&messageBar.text)
}
