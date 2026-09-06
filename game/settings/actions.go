package settings

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/go-gl/glfw/v3.3/glfw"
	"tophatdemon.com/total-invasion-ii/engine/input"
)

type Action [2]input.Binding

var (
	// Cheat codes
	ActionNoclip  = input.NewCharSequenceBinding(glfw.KeyT, glfw.KeyD, glfw.KeyC, glfw.KeyL, glfw.KeyI, glfw.KeyP) //TDCLIP
	ActionGodMode = input.NewCharSequenceBinding(glfw.KeyT, glfw.KeyD, glfw.KeyD, glfw.KeyQ, glfw.KeyD)            //TDDQD
	ActionMarySue = input.NewCharSequenceBinding(glfw.KeyT, glfw.KeyD, glfw.KeyM, glfw.KeyS, glfw.KeyM)            //TDMSM
	ActionDie     = input.NewCharSequenceBinding(
		glfw.KeyT, glfw.KeyD, glfw.KeyU, glfw.KeyN, glfw.KeyA, glfw.KeyL, glfw.KeyI, glfw.KeyV, glfw.KeyE,
	) //TDUNALIVE
	ActionKillEnemies = input.NewCharSequenceBinding(
		glfw.KeyT, glfw.KeyD, glfw.KeyN, glfw.KeyU, glfw.KeyK, glfw.KeyE, glfw.KeyM,
	) //TDNUKEM
	ActionCastBlessing = input.NewCharSequenceBinding(
		glfw.KeyT, glfw.KeyD, glfw.KeyW, glfw.KeyO, glfw.KeyL, glfw.KeyO, glfw.KeyL, glfw.KeyO,
	) //TDWOLOLO
	ActionLaunchEditor = input.NewCharSequenceBinding(glfw.KeyT, glfw.KeyD, glfw.KeyJ, glfw.KeyO, glfw.KeyM, glfw.KeyT) //TDJOMT
	ActionSpawnChicken = input.NewCharSequenceBinding(glfw.KeyT, glfw.KeyD, glfw.KeyK, glfw.KeyF, glfw.KeyC)            //TDKFC
	ActionLevelSelect  = input.NewCharSequenceBinding(glfw.KeyT, glfw.KeyD, glfw.KeyC, glfw.KeyL, glfw.KeyE, glfw.KeyV) //TDCLEV
)

func (action Action) Axis() float32 {
	for _, binding := range action {
		if binding != nil {
			if axis := binding.Axis(); axis != 0.0 {
				return axis
			}
		}
	}
	return 0.0
}

func (action Action) Pressed() bool {
	for _, binding := range action {
		if binding != nil && binding.Pressed() {
			return true
		}
	}
	return false
}

func (action Action) JustPressed() bool {
	for _, binding := range action {
		if binding != nil && binding.JustPressed() {
			return true
		}
	}
	return false
}

func (action Action) JustReleased() bool {
	for _, binding := range action {
		if binding != nil && binding.JustReleased() {
			return true
		}
	}
	return false
}

func (action Action) LocalizationKey() string {
	for _, binding := range action {
		if binding != nil {
			return binding.LocalizationKey()
		}
	}
	return "???"
}

func (action Action) String() string {
	return Localize(action.LocalizationKey())
}

func (action Action) MarshalTOML() ([]byte, error) {
	var builder strings.Builder
	builder.WriteString("[\n")
	encoder := toml.NewEncoder(&builder)
	for _, binding := range action {
		if binding != nil {
			builder.WriteRune('\t')
			err := encoder.Encode(binding)
			if err != nil {
				return nil, err
			}
			builder.WriteString(",\n")
		}
	}
	builder.WriteRune(']')
	return []byte(builder.String()), nil
}

func (action *Action) UnmarshalTOML(data any) error {
	clear(action[:])
	dataSlice, ok := data.([]any)
	if !ok {
		return fmt.Errorf("action must be an array")
	}
	if len(dataSlice) > len(action) {
		return fmt.Errorf("too many bindings (%v) assigned to an action; maximum is %v", len(dataSlice), len(action))
	}
	for i, bindingData := range dataSlice {
		bindingMap, isMap := bindingData.(map[string]any)
		if !isMap || bindingMap == nil {
			return fmt.Errorf("binding at index %v should be a map", i)
		}
		if value, hasKey := bindingMap["Key"]; hasKey {
			if keyNumber, ok := value.(int64); ok {
				action[i] = input.NewKeyBinding(glfw.Key(keyNumber))
			} else {
				return fmt.Errorf("binding Key value must be an integer")
			}
		} else if value, hasMouseButton := bindingMap["MouseButton"]; hasMouseButton {
			if buttonNumber, ok := value.(int64); ok {
				action[i] = input.NewMouseButtonBinding(glfw.MouseButton(buttonNumber))
			} else {
				return fmt.Errorf("binding MouseButton value must be an integer")
			}
		} else if value, hasMouseAxis := bindingMap["MouseAxis"]; hasMouseAxis {
			var axisNumber int64
			var ok bool
			if axisNumber, ok = value.(int64); !ok {
				return fmt.Errorf("binding MouseAxis value must be an integer")
			}
			action[i] = input.NewMouseMovementBinding(input.MouseAxis(axisNumber))
		} else {
			return fmt.Errorf("this type of binding is unsupported")
		}
	}
	return nil
}
