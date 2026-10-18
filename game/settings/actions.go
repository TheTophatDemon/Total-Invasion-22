package settings

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/rgfw"
)

type Action [2]input.Binding

var (
	// Cheat codes
	ActionNoclip  = input.NewCharSequenceBinding(rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_C, rgfw.KEY_L, rgfw.KEY_I, rgfw.KEY_P) //TDCLIP
	ActionGodMode = input.NewCharSequenceBinding(rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_D, rgfw.KEY_Q, rgfw.KEY_D)             //TDDQD
	ActionMarySue = input.NewCharSequenceBinding(rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_M, rgfw.KEY_S, rgfw.KEY_M)             //TDMSM
	ActionDie     = input.NewCharSequenceBinding(
		rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_U, rgfw.KEY_N, rgfw.KEY_A, rgfw.KEY_L, rgfw.KEY_I, rgfw.KEY_V, rgfw.KEY_E,
	) //TDUNALIVE
	ActionKillEnemies = input.NewCharSequenceBinding(
		rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_N, rgfw.KEY_U, rgfw.KEY_K, rgfw.KEY_E, rgfw.KEY_M,
	) //TDNUKEM
	ActionCastBlessing = input.NewCharSequenceBinding(
		rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_W, rgfw.KEY_O, rgfw.KEY_L, rgfw.KEY_O, rgfw.KEY_L, rgfw.KEY_O,
	) //TDWOLOLO
	ActionLaunchEditor = input.NewCharSequenceBinding(rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_J, rgfw.KEY_O, rgfw.KEY_M, rgfw.KEY_T) //TDJOMT
	ActionSpawnChicken = input.NewCharSequenceBinding(rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_K, rgfw.KEY_F, rgfw.KEY_C)             //TDKFC
	ActionLevelSelect  = input.NewCharSequenceBinding(rgfw.KEY_T, rgfw.KEY_D, rgfw.KEY_C, rgfw.KEY_L, rgfw.KEY_E, rgfw.KEY_V) //TDCLEV
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
				action[i] = input.NewKeyBinding(rgfw.Key(keyNumber))
			} else {
				return fmt.Errorf("binding Key value must be an integer")
			}
		} else if value, hasMouseButton := bindingMap["MouseButton"]; hasMouseButton {
			if buttonNumber, ok := value.(int64); ok {
				action[i] = input.NewMouseButtonBinding(rgfw.MouseButton(buttonNumber))
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
