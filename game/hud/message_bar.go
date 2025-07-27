package hud

import (
	"unicode/utf8"

	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type messageBar struct {
	text       ui.Text
	background ui.Box
	timer      float32
	priority   int
	flash      float32 // Tracks the color change in the message bar after a message is shown
}

func (messageBar *messageBar) init() {
	messageBar.background = ui.Box{
		Color: color.Black,
		Src: math2.Rect{
			Width: 1.0, Height: 1.0,
		},
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      0.0,
				Y:      settings.UIHeight() - 32.0,
				Width:  settings.UIWidth(),
				Height: 32.0,
			},
			Depth: 2.0,
		},
	}

	messageBar.text = ui.Text{
		Settings: ui.TextSettings{
			WrapWords: false,
		},
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      messageBar.background.Dest.X + 8.0,
				Y:      messageBar.background.Dest.Y + 2.0,
				Width:  messageBar.background.Dest.Width - 16.0,
				Height: messageBar.background.Dest.Height - 2.0,
			},
			Depth: 3.0,
			Scale: 1.0,
		},
	}
}

func (messageBar *messageBar) ShowMessage(text string, duration float32, priority int, colr color.Color) {
	if priority >= messageBar.priority {
		messageBar.timer = duration
		messageBar.priority = priority
		messageBar.text.SetText(text)
		messageBar.text.Color = colr
		messageBar.flash = 0.5
	}
}

func (messageBar *messageBar) layout(queue *ui.RenderQueue, deltaTime float32) {
	messageBar.timer -= deltaTime
	if messageBar.timer <= 0.0 {
		const scrollSpeed = -0.1
		if messageBar.timer < scrollSpeed {
			messageBar.timer = 0.0
			msgText := messageBar.text.Text()
			if len(msgText) > 1 {
				_, byteCount := utf8.DecodeRuneInString(msgText)
				messageBar.text.SetText(msgText[byteCount:])
			} else {
				messageBar.priority = 0
				messageBar.text.Color = color.Transparent
				messageBar.text.SetText("")
			}
		}
	}
	messageBar.flash = max(0.0, messageBar.flash-deltaTime)
	messageBar.background.Color = color.Color{
		R: (1.0 - messageBar.text.Color.R) * messageBar.flash,
		G: (1.0 - messageBar.text.Color.G) * messageBar.flash,
		B: (1.0 - messageBar.text.Color.B) * messageBar.flash,
		A: 1.0,
	}

	queue.Add(&messageBar.text)
	queue.Add(&messageBar.background)
}
