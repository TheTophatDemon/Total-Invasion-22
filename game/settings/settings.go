package settings

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/go-gl/glfw/v3.3/glfw"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
)

type (
	Locale string
	Action [2]input.Binding
	Data   struct {
		WindowWidth, WindowHeight uint16
		Fullscreen, Vsync         bool
		MouseSensitivity          int
		TextShadowTransparency    float32 // From 0 to 1
		SfxVolume, MusicVolume    float32 // From 0 to 1
		Locale                    Locale
		Fov                       float32 // Measured in degrees
		ChickenHarm               bool    // Allow harm to chickens
		Debug                     struct {
			StartMap string
		}
		DifficultyIndex int

		// Menu actions
		ActionMenuDown, ActionMenuUp, ActionMenuConfirm            Action
		ActionMenuCancel, ActionMenuIncrement, ActionMenuDecrement Action

		// In game actions
		ActionForward, ActionBack, ActionLeft, ActionRight Action
		ActionSlow, ActionLookLeft, ActionLookRight        Action
		ActionFire, ActionAltFire, ActionUse               Action

		// Weapon selection actions
		ActionSickle, ActionChicken, ActionGrenade, ActionParusu         Action
		ActionDblGrenade, ActionSign, ActionAirhorn, ActionDefenestrator Action
		ActionCluckster, ActionWeaponWheel                               Action
	}
)

const (
	LocaleEnglish    Locale = "en"
	LocaleRussian    Locale = "ru"
	settingsFilePath        = "game_settings.toml"
)

var (
	Default, Current Data

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

	scaledMouseSensitivities = [9]float32{
		0.001,
		0.002,
		0.003,
		0.004,
		0.005,
		0.007,
		0.009,
		0.013,
		0.020,
	}
)

func init() {
	Default = Data{
		WindowWidth: 1280, WindowHeight: 720,
		Fullscreen:             false,
		Vsync:                  true,
		MouseSensitivity:       5,
		TextShadowTransparency: 0.5,
		SfxVolume:              1.0, MusicVolume: 1.0,
		Locale:              LocaleEnglish,
		Fov:                 70.0,
		ChickenHarm:         true,
		DifficultyIndex:     len(Difficulties) - 1,
		ActionMenuDown:      Action{input.NewKeyBinding(glfw.KeyDown), input.NewKeyBinding(glfw.KeyS)},
		ActionMenuUp:        Action{input.NewKeyBinding(glfw.KeyUp), input.NewKeyBinding(glfw.KeyW)},
		ActionMenuConfirm:   Action{input.NewKeyBinding(glfw.KeyEnter)},
		ActionMenuCancel:    Action{input.NewKeyBinding(glfw.KeyEscape)},
		ActionMenuIncrement: Action{input.NewKeyBinding(glfw.KeyRight), input.NewKeyBinding(glfw.KeyD)},
		ActionMenuDecrement: Action{input.NewKeyBinding(glfw.KeyLeft), input.NewKeyBinding(glfw.KeyA)},
		ActionForward:       Action{input.NewKeyBinding(glfw.KeyW)},
		ActionBack:          Action{input.NewKeyBinding(glfw.KeyS)},
		ActionLeft:          Action{input.NewKeyBinding(glfw.KeyA)},
		ActionRight:         Action{input.NewKeyBinding(glfw.KeyD)},
		ActionSlow:          Action{input.NewKeyBinding(glfw.KeyLeftShift), input.NewKeyBinding(glfw.KeyRightShift)},
		ActionLookLeft:      Action{input.NewMouseMovementBinding(input.MouseAxisNegX, ScaledMouseSensitivity(5)), input.NewKeyBinding(glfw.KeyLeft)},
		ActionLookRight:     Action{input.NewMouseMovementBinding(input.MouseAxisPosX, ScaledMouseSensitivity(5)), input.NewKeyBinding(glfw.KeyRight)},
		ActionFire:          Action{input.NewMouseButtonBinding(glfw.MouseButtonLeft), input.NewKeyBinding(glfw.KeyLeftControl)},
		ActionAltFire:       Action{input.NewMouseButtonBinding(glfw.MouseButtonRight), input.NewKeyBinding(glfw.KeyLeftAlt)},
		ActionUse:           Action{input.NewKeyBinding(glfw.KeyE)},
		ActionSickle:        Action{input.NewKeyBinding(glfw.Key1)},
		ActionChicken:       Action{input.NewKeyBinding(glfw.Key2)},
		ActionGrenade:       Action{input.NewKeyBinding(glfw.Key3)},
		ActionParusu:        Action{input.NewKeyBinding(glfw.Key4)},
		ActionDblGrenade:    Action{input.NewKeyBinding(glfw.Key5)},
		ActionSign:          Action{input.NewKeyBinding(glfw.Key6)},
		ActionAirhorn:       Action{input.NewKeyBinding(glfw.Key7)},
		ActionDefenestrator: Action{input.NewKeyBinding(glfw.Key8)},
		ActionCluckster:     Action{input.NewKeyBinding(glfw.Key9)},
		ActionWeaponWheel:   Action{input.NewKeyBinding(glfw.KeyQ)},
	}
	Current = Default
}

func (data *Data) WindowAspectRatio() float32 {
	return float32(data.WindowWidth) / float32(data.WindowHeight)
}

func LoadOrInit() {
	settingsFile, err := os.Open(settingsFilePath)
	if errors.Is(err, os.ErrNotExist) {
		Save()
	} else if err != nil {
		failure.LogErrWithLocation("Could not open settings file; %v", err)
		return
	} else {
		defer settingsFile.Close()

		fileBytes, err := io.ReadAll(settingsFile)
		if err != nil {
			failure.LogErrWithLocation("Could not read settings file; %v", err)
			return
		}

		_, err = toml.Decode(string(fileBytes), &Current)
		if err != nil {
			failure.LogErrWithLocation("Could not unmarshal settings file; %v", err)
			Current = Default
			return
		}
	}
}

func Save() {
	settingsFile, err := os.Create(settingsFilePath)
	if err != nil {
		failure.LogErrWithLocation("Could not open settings file for writing; %v", err)
	}
	defer settingsFile.Close()

	settingsBytes, err := toml.Marshal(Current)
	if err != nil {
		failure.LogErrWithLocation("Could not marshal settings; %v", err)
		return
	}

	_, err = settingsFile.Write(settingsBytes)

	if err != nil {
		failure.LogErrWithLocation("Could not write settings to file; %v", err)
		return
	}
}

func Localize(key string) string {
	return LocalizeWith(key, Current.Locale, "")
}

func LocalizeWith(key string, locale Locale, grammarCase string) string {
	trans, err := cache.GetTranslation(fmt.Sprintf("assets/translations/strings_%v.toml", string(locale)))
	if err != nil {
		failure.LogErrWithLocation("failed to retrieve strings in %v for key %v: %v", locale, key, err)
		return "ERROR"
	}
	localizedText, ok := (*trans)[key+grammarCase]
	if !ok && len(grammarCase) > 0 {
		// When the key is not found with this grammar case, use the key by itself.
		localizedText, ok = (*trans)[key]
	}
	if !ok {
		// Fall back to English
		trans, err = cache.GetTranslation(fmt.Sprintf("assets/translations/strings_%v.toml", string(LocaleEnglish)))
		if err != nil {
			failure.LogErrWithLocation("failed to retrieve English fallback for localization key %v: %v", key, err)
			return "ERROR"
		}
		localizedText = (*trans)[key]
		if localizedText == "" {
			// There's no English translation, so it should show the key verbatim instead.
			// This is needed for things like Keyboard key bindings to show up correctly.
			return key
		}
	}

	// Parse as a text template in order to substitute control names and such.
	templ, err := template.New(key).
		Funcs(template.FuncMap{
			"acc": func(input any) string {
				switch in := input.(type) {
				case Action:
					return LocalizeWith(in.LocalizationKey(), locale, "Accusative")
				case string:
					return LocalizeWith(in, locale, "Accusative")
				}
				return "ERROR"
			},
		}).
		Parse(localizedText)

	if err != nil {
		failure.LogErrWithLocation("failed to parse template for key %v in lang %v: %v", key, locale, err)
		return "ERROR"
	}

	var finalText strings.Builder
	err = templ.Execute(&finalText, Current)
	if err != nil {
		failure.LogErrWithLocation("failed to execute template for key %v in lang %v: %v", key, locale, err)
		return "ERROR"
	}

	return finalText.String()
}

// Deprecated: UI Elements should not normally be scaled with the screen size.
func UIScale() float32 {
	return float32(Current.WindowHeight) / 480
}

// Returns the size the sprites on the HUD should be scaled to.
func SpriteScale() float32 {
	return float32(Current.WindowHeight) / 240
}

func UIWidth() float32 {
	return float32(Current.WindowWidth)
}

func UIHeight() float32 {
	return float32(Current.WindowHeight)
}

// Returns a floating point value for the given mouse sensitivity value that should be applied to the mouse movement.
func ScaledMouseSensitivity(intSensitivity int) float32 {
	return scaledMouseSensitivities[max(0, min(len(scaledMouseSensitivities)-1, intSensitivity-1))]
}

func (locale Locale) String() string {
	return LocalizeWith("myLang", locale, "")
}

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
			return binding.String()
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
			var sensitivity float64
			var ok bool
			if axisNumber, ok = value.(int64); !ok {
				return fmt.Errorf("binding MouseAxis value must be an integer")
			}
			var hasSensitivity bool
			value, hasSensitivity = bindingMap["Sensitivity"]
			if !hasSensitivity {
				return fmt.Errorf("binding is missing Sensitivity value")
			}
			if sensitivity, ok = value.(float64); !ok {
				return fmt.Errorf("binding Sensitivity value must be a float")
			}
			action[i] = input.NewMouseMovementBinding(input.MouseAxis(axisNumber), float32(sensitivity))
		} else {
			return fmt.Errorf("this type of binding is unsupported")
		}
	}
	return nil
}
