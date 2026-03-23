package screens

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const sfxSlide = "assets/sounds/ui/stats_ding.wav"

type Slider struct {
	MenuItem
	txtLabel, txtLeft, boxSlider, boxKnob, txtRight, txtValue ui.Element
	min, max, step, count                                     float64
	value                                                     float64
	grabbed                                                   bool
}

func (sl *Slider) Init(labelKey string, min, max, step, initialValue float64) *Slider {
	*sl = Slider{
		min:   min,
		max:   max,
		step:  step,
		value: initialValue,
		count: math2.Floor((1.0 + max - min) / step),
	}

	smallScreen := settings.Current.WindowWidth <= 800

	sl.txtLabel = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
	}, settings.Localize(labelKey)+": ", ui.TextConfig{
		Color:     maybe.Some(color.White),
		WrapWords: true,
	})
	sl.txtLabel.FitText()
	if smallScreen && sl.txtLabel.Size()[0] >= 300.0 {
		sl.txtLabel.SetSize(mgl32.Vec2{
			300.0,
			64.0,
		})
	}

	sl.txtLeft = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
		Size:   mgl32.Vec2{16, 24},
	}, "<", ui.TextConfig{
		Color: maybe.Some(color.Yellow),
	})

	sl.boxSlider = ui.NewBox(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
		Size:   mgl32.Vec2{128, 8},
	}, cache.GetTexture("assets/textures/ui/slider_background.png"))

	knobTex := cache.GetTexture("assets/textures/ui/slider_knob.png")
	sl.boxKnob = ui.NewBox(ui.Transform{
		Origin: ui.Ratios{0.5, 0.5},
		Depth:  11,
		Size:   mgl32.Vec2{24.0, 24.0},
	}, knobTex)
	sl.boxKnob.AnimPlayer.ChangeAnimation(knobTex.GetDefaultAnimation())

	sl.txtRight = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
		Size:   mgl32.Vec2{16, 24},
	}, ">", ui.TextConfig{
		Color: maybe.Some(color.Yellow),
	})

	sl.txtValue = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
		Size:   mgl32.Vec2{80, 24},
	}, fmt.Sprintf("%v", initialValue), ui.TextConfig{})

	return sl
}

func (sl *Slider) next() {
	cache.GetSfx(sfxSlide).Play()
	snappedValue := math2.Floor(sl.value/sl.step) * sl.step
	sl.value = min(snappedValue+sl.step, sl.max)
}

func (sl *Slider) prev() {
	cache.GetSfx(sfxSlide).Play()
	snappedValue := math2.Ceil(sl.value/sl.step) * sl.step
	sl.value = max(sl.min, snappedValue-sl.step)
}

func (sl *Slider) Input(action MenuInputType, menu *Menu) {
	sl.MenuItem.Input(action, menu)
	switch action {
	case MenuInputIncrement:
		sl.next()
	case MenuInputDecrement:
		sl.prev()
	case MenuInputClick:
		mousePos := input.MousePosition()
		if sl.txtLeft.OnScreenBox().ContainsPoint(mousePos) {
			sl.prev()
		} else if sl.txtRight.OnScreenBox().ContainsPoint(mousePos) {
			sl.next()
		}
	}
}

func (sl *Slider) Focus() {
	sl.MenuItem.Focus()
	configMaybe := sl.txtLabel.TextConfig()
	if config, ok := configMaybe.Get(); ok {
		config.Color = maybe.Some(color.Yellow)
	}
}

func (sl *Slider) Blur() {
	sl.MenuItem.Blur()
	configMaybe := sl.txtLabel.TextConfig()
	if config, ok := configMaybe.Get(); ok {
		config.Color = maybe.Some(color.White)
	}
}

func (sl *Slider) Layout(queue *ui.RenderQueue, deltaTime float32) {
	// Slide when mouse is held down
	mousePos := input.MousePosition()
	sliderBox := sl.boxSlider.OnScreenBox()
	clickableBox := math2.Rect{
		X:      sliderBox.X - sl.boxKnob.Width()/2.0,
		Y:      sliderBox.Y - sl.boxKnob.Height()/2.0,
		Width:  sliderBox.Width + sl.boxKnob.Width(),
		Height: sliderBox.Height + sl.boxKnob.Height(),
	}
	if input.AnyMouseButtonJustPressed() {
		if clickableBox.ContainsPoint(mousePos) {
			sl.grabbed = true
		}
	} else if !input.AnyMouseButtonPressed() {
		sl.grabbed = false
	}
	if sl.grabbed {
		offset := float64((mousePos[0] - sliderBox.X) / sliderBox.Width)
		// prev := sl.value
		sl.value = min(sl.max, max(sl.min, sl.min+((sl.max-sl.min)*offset)))
		// if !mgl64.FloatEqual(prev, sl.value) {
		// 	cache.GetSfx(sfxSlide).Play()
		// }
		// TODO: Make this sound less obnoxious
	}

	sl.txtValue.SetText(fmt.Sprintf("%.2f", sl.value))

	components := []*ui.Element{&sl.txtLabel, &sl.txtLeft, &sl.boxSlider, &sl.txtRight, &sl.txtValue}
	ui.LayoutStack(&sl.element, ui.StackParams{
		Gap: 8.0,
	}, components...)
	queue.Add(components...)

	sl.boxKnob.SetPosition(sl.boxSlider.Position().Add(mgl32.Vec2{float32(sl.FractionValue()) * sl.boxSlider.Width()}))
	sl.boxKnob.AnimPlayer.MoveToFrame(int(sl.FractionValue() * float64(len(sl.boxKnob.AnimPlayer.CurrentAnimation().Frames)-1)))
	queue.Add(&sl.boxKnob)

	sl.MenuItem.Layout(queue, deltaTime)
}

func (sl *Slider) IntValue() int {
	return int(sl.value)
}

func (sl *Slider) FloatValue() float64 {
	return sl.value
}

func (sl *Slider) FractionValue() float64 {
	return (sl.value - sl.min) / (sl.max - sl.min)
}
