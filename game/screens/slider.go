package screens

import (
	"fmt"
	"strings"

	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const sfxSlide = "assets/sounds/ui/stats_ding.wav"

type Slider struct {
	MenuItem
	txtLabel, txtLeft, txtSlider, txtRight, txtValue ui.Element
	min, max, step, count                            int
	value                                            int
}

func (sl *Slider) Init(labelKey string, min, max, step, initialValue int) *Slider {
	*sl = Slider{
		min:   min,
		max:   max,
		step:  step,
		value: initialValue,
		count: (1 + max - min) / step,
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

	sl.txtSlider = ui.NewText(ui.Transform{
		Origin: ui.Ratios{0.0, 0.5},
		Depth:  10,
	}, strings.Repeat("_", sl.count+1), ui.TextConfig{
		Align: ui.TextAlignCenterH,
	})
	sl.txtSlider.FitText()

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
	sl.value = min(sl.value+sl.step, sl.max)
}

func (sl *Slider) prev() {
	cache.GetSfx(sfxSlide).Play()
	sl.value = max(sl.min, sl.value-sl.step)
}

func (sl *Slider) Input(action MenuInputType) {
	sl.MenuItem.Input(action)
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
		sl.txtLabel.SetTextConfig(*config)
	}
}

func (sl *Slider) Blur() {
	sl.MenuItem.Blur()
	configMaybe := sl.txtLabel.TextConfig()
	if config, ok := configMaybe.Get(); ok {
		config.Color = maybe.Some(color.White)
		sl.txtLabel.SetTextConfig(*config)
	}
}

func (sl *Slider) Layout(queue *ui.RenderQueue, deltaTime float32) {
	// Slide when mouse is held down
	mousePos := input.MousePosition()
	if input.IsMouseButtonDown(glfw.MouseButton1) || input.IsMouseButtonDown(glfw.MouseButton2) {
		if sl.txtSlider.OnScreenBox().ContainsPoint(mousePos) {
			offset := (mousePos[0] - sl.txtSlider.Position()[0]) / sl.txtSlider.Size()[0]
			prev := sl.value
			sl.value = min(sl.max, sl.min+int(float32(sl.count+1)*offset)*sl.step)
			if prev != sl.value {
				cache.GetSfx(sfxSlide).Play()
			}
		}
	}

	var sliderBuilder strings.Builder
	sliderBuilder.Grow(sl.count)
	for i := sl.min; i <= sl.max; i += sl.step {
		if i == sl.value {
			sliderBuilder.WriteRune('*')
		} else {
			sliderBuilder.WriteRune('_')
		}
	}
	sl.txtSlider.SetText(sliderBuilder.String())
	sl.txtValue.SetText(fmt.Sprintf("%v", sl.value))

	components := []*ui.Element{&sl.txtLabel, &sl.txtLeft, &sl.txtSlider, &sl.txtRight, &sl.txtValue}
	ui.LayoutStack(&sl.element, ui.StackParams{
		Gap: 8.0,
	}, components...)
	queue.Add(components...)

	sl.MenuItem.Layout(queue, deltaTime)
}

func (sl *Slider) IntValue() int {
	return sl.value
}

func (sl *Slider) FractionValue() float32 {
	return float32(sl.value-sl.min) / float32(sl.max-sl.min)
}
