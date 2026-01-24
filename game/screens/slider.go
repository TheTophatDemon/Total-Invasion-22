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
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const sfxSlide = "assets/sounds/ui/stats_ding.wav"

type Slider struct {
	MenuItem
	txtLeft, txtSlider, txtRight, txtValue ui.Element
	labelWidth                             float32
	min, max, step, count                  int
	value                                  int
}

func (sl *Slider) Init(labelKey string, min, max, step, initialValue int) *Slider {
	*sl = Slider{
		min:   min,
		max:   max,
		step:  step,
		value: initialValue,
		count: (1 + max - min) / step,
	}
	sl.MenuItem = MenuItem{
		element: ui.NewText(ui.Transform{}, settings.Localize(labelKey)+": ", ui.TextConfig{}),
	}
	sl.MenuItem.FitText()

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

	sl.labelWidth = sl.OnScreenBox().Width

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

func (sl *Slider) Input(action input.Action) {
	sl.MenuItem.Input(action)
	switch action {
	case settings.ActionMenuIncrement:
		sl.next()
	case settings.ActionMenuDecrement:
		sl.prev()
	case settings.ActionMenuClick:
		mousePos := input.MousePosition()
		if sl.txtLeft.OnScreenBox().ContainsPoint(mousePos) {
			sl.prev()
		} else if sl.txtRight.OnScreenBox().ContainsPoint(mousePos) {
			sl.next()
		}
	}
}

func (sl *Slider) Layout(queue *ui.RenderQueue, deltaTime float32) {
	// Slide when mouse is held down
	mousePos := input.MousePosition()
	if input.IsMouseButtonDown(glfw.MouseButton1) || input.IsMouseButtonDown(glfw.MouseButton2) {
		if sl.txtSlider.OnScreenBox().ContainsPoint(mousePos) {
			offset := (mousePos[0] - sl.txtSlider.Position()[0]) / sl.txtSlider.Size()[0]
			prev := sl.value
			sl.value = sl.min + int(math2.Round(float32(sl.count)*offset))*sl.step
			if prev != sl.value {
				cache.GetSfx(sfxSlide).Play()
			}
		}
	}

	sl.txtLeft.SetPosition(sl.Position())
	sl.txtLeft.Translate(mgl32.Vec2{sl.labelWidth, 0.0})
	queue.Add(&sl.txtLeft)

	leftWidth := sl.txtLeft.OnScreenBox().Width + 8.0
	sl.txtSlider.SetPosition(sl.txtLeft.Position())

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

	sl.txtSlider.Translate(mgl32.Vec2{leftWidth, 0.0})
	queue.Add(&sl.txtSlider)

	valueWidth := sl.txtSlider.OnScreenBox().Width + 8.0
	sl.txtRight.SetPosition(sl.txtSlider.Position())
	sl.txtRight.Translate(mgl32.Vec2{valueWidth, 0.0})
	queue.Add(&sl.txtRight)

	sl.txtValue.SetPosition(mgl32.Vec2{
		sl.txtRight.Position()[0] + sl.txtRight.OnScreenBox().Width + 8.0,
		sl.txtRight.Position()[1],
	})
	sl.txtValue.SetText(fmt.Sprintf("%v", sl.value))
	queue.Add(&sl.txtValue)

	clickableWidth := sl.labelWidth + leftWidth + valueWidth + sl.txtRight.OnScreenBox().Width
	sl.SetSize(mgl32.Vec2{clickableWidth, sl.Size()[1]})

	sl.MenuItem.Layout(queue, deltaTime)
}

func (sl *Slider) IntValue() int {
	return sl.value
}

func (sl *Slider) FractionValue() float32 {
	return float32(sl.value-sl.min) / float32(sl.max-sl.min)
}
