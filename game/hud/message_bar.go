package hud

import (
	"unicode/utf8"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type MessageBar struct {
	text     ui.Element
	priority int
	flash    gween.Tween    // Tracks the color change in the message bar after a message is shown
	scroll   gween.Sequence // Tracks the appearance of the text as it scrolls off screen.
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
		scrollTweens := make([]*gween.Tween, 0, charCount+1)
		// Pause to let the player read
		scrollTweens = append(scrollTweens, gween.New(0.0, 0.0, float32(charCount)*(3.0/80.0), ease.Linear))
		// Adds a tween for removing each character in sequence
		for i := range charCount {
			scrollTweens = append(scrollTweens, gween.New(float32(i), float32(i+1), 0.1, ease.Linear))
		}
		messageBar.scroll = *gween.NewSequence(scrollTweens...)

		messageBar.priority = priority
		messageBar.text.SetText(text)
		messageBar.text.SetTextConfig(ui.TextConfig{
			Color:         maybe.Some(colr),
			WrapWords:     false,
			DisableShadow: true,
		})
		messageBar.flash = *gween.New(1.0, 0.0, 0.5, ease.OutCubic)
	}
}

func (messageBar *MessageBar) layout(queue *ui.RenderQueue, deltaTime float32) {
	_, shouldScroll, _ := messageBar.scroll.Update(deltaTime)
	if shouldScroll {
		msgText := messageBar.text.Text()
		if len(msgText) > 1 {
			_, byteCount := utf8.DecodeRuneInString(msgText)
			messageBar.text.SetText(msgText[byteCount:])
		} else {
			messageBar.priority = 0
			messageBar.text.SetText("")
		}
	}

	flashAmt, _ := messageBar.flash.Update(deltaTime)
	txtColor := messageBar.text.TextConfig().Unwrap().Color.Or(color.White)
	messageBar.text.BgColor = maybe.Some(color.Color{
		R: (1.0 - txtColor.R) * flashAmt,
		G: (1.0 - txtColor.G) * flashAmt,
		B: (1.0 - txtColor.B) * flashAmt,
		A: 1.0,
	})

	queue.Add(&messageBar.text)
}
