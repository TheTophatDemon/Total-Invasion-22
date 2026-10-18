package settings

import (
	"errors"
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/rgfw"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
)

type (
	Data struct {
		WindowWidth, WindowHeight uint16
		Fullscreen, Vsync         bool
		MouseSensitivity          float64
		TextShadowTransparency    float64 // From 0 to 1
		SfxVolume, MusicVolume    float64 // From 0 to 1
		Locale                    Locale
		DifficultyIndex           int     `json:"-"` // Index into Difficulties array.
		Fov                       float64 // Measured in degrees
		ChickenHarm               bool    // Allow harm to chickens
		Debug                     struct {
			StartMap string
		}
		Pixelization int
		// Menu actions
		ActionMenuDown      Action
		ActionMenuUp        Action
		ActionMenuConfirm   Action
		ActionMenuCancel    Action
		ActionMenuIncrement Action
		ActionMenuDecrement Action

		// In game actions
		ActionForward, ActionBack, ActionLeft, ActionRight                                   Action
		ActionSlow, ActionLookLeft, ActionLookRight, ActionFastLookLeft, ActionFastLookRight Action
		ActionFire, ActionAltFire, ActionUse                                                 Action

		// Weapon selection actions
		ActionSickle, ActionChicken, ActionGrenade, ActionParusu         Action
		ActionDblGrenade, ActionSign, ActionAirhorn, ActionDefenestrator Action
		ActionCluckster, ActionWeaponWheel                               Action
	}
)

const settingsFilePath = "game_settings.toml"

var (
	Default, Current Data
)

func init() {
	Default = Data{
		WindowWidth: 1280, WindowHeight: 720,
		Pixelization:           2,
		Fullscreen:             false,
		Vsync:                  true,
		DifficultyIndex:        1,
		MouseSensitivity:       6.0,
		TextShadowTransparency: 0.5,
		SfxVolume:              1.0, MusicVolume: 1.0,
		Locale:              LocaleEnglish,
		Fov:                 70.0,
		ChickenHarm:         true,
		ActionMenuDown:      Action{input.NewKeyBinding(rgfw.KEY_DOWN), input.NewKeyBinding(rgfw.KEY_S)},
		ActionMenuUp:        Action{input.NewKeyBinding(rgfw.KEY_UP), input.NewKeyBinding(rgfw.KEY_W)},
		ActionMenuConfirm:   Action{input.NewKeyBinding(rgfw.KEY_ENTER)},
		ActionMenuCancel:    Action{input.NewKeyBinding(rgfw.KEY_ESCAPE)},
		ActionMenuIncrement: Action{input.NewKeyBinding(rgfw.KEY_RIGHT), input.NewKeyBinding(rgfw.KEY_D)},
		ActionMenuDecrement: Action{input.NewKeyBinding(rgfw.KEY_LEFT), input.NewKeyBinding(rgfw.KEY_A)},
		ActionForward:       Action{input.NewKeyBinding(rgfw.KEY_W)},
		ActionBack:          Action{input.NewKeyBinding(rgfw.KEY_S)},
		ActionLeft:          Action{input.NewKeyBinding(rgfw.KEY_A)},
		ActionRight:         Action{input.NewKeyBinding(rgfw.KEY_D)},
		ActionSlow:          Action{input.NewKeyBinding(rgfw.KEY_SHIFT_L), input.NewKeyBinding(rgfw.KEY_SHIFT_R)},
		ActionLookLeft:      Action{input.NewMouseMovementBinding(input.MouseAxisNegX), input.NewKeyBinding(rgfw.KEY_LEFT)},
		ActionFastLookLeft:  Action{nil, nil},
		ActionLookRight:     Action{input.NewMouseMovementBinding(input.MouseAxisPosX), input.NewKeyBinding(rgfw.KEY_RIGHT)},
		ActionFastLookRight: Action{nil, nil},
		ActionFire:          Action{input.NewMouseButtonBinding(rgfw.MOUSE_BUTTON_LEFT), input.NewKeyBinding(rgfw.KEY_CONTROL_L)},
		ActionAltFire:       Action{input.NewMouseButtonBinding(rgfw.MOUSE_BUTTON_RIGHT), input.NewKeyBinding(rgfw.KEY_ALT_L)},
		ActionUse:           Action{input.NewKeyBinding(rgfw.KEY_E)},
		ActionSickle:        Action{input.NewKeyBinding(rgfw.KEY0)},
		ActionChicken:       Action{input.NewKeyBinding(rgfw.KEY2)},
		ActionGrenade:       Action{input.NewKeyBinding(rgfw.KEY3)},
		ActionParusu:        Action{input.NewKeyBinding(rgfw.KEY4)},
		ActionDblGrenade:    Action{input.NewKeyBinding(rgfw.KEY5)},
		ActionSign:          Action{input.NewKeyBinding(rgfw.KEY6)},
		ActionAirhorn:       Action{input.NewKeyBinding(rgfw.KEY7)},
		ActionDefenestrator: Action{input.NewKeyBinding(rgfw.KEY8)},
		ActionCluckster:     Action{input.NewKeyBinding(rgfw.KEY9)},
		ActionWeaponWheel:   Action{input.NewKeyBinding(rgfw.KEY_Q)},
	}
	Current = Default
}

func (data *Data) WindowAspectRatio() float32 {
	return float32(data.WindowWidth) / float32(data.WindowHeight)
}

// Applies the settings to the state of the engine, UI, and audio
func (data Data) Apply() {
	engine.SetVideoMode(data.Fullscreen, int(data.WindowWidth), int(data.WindowHeight))
	ui.GlobalTextScale = data.TextScale()
	ui.SetTextShadowColor(color.Black.WithAlpha(float32(data.TextShadowTransparency)))

	engine.SetVsync(data.Vsync)
	tdaudio.SetSfxVolume(float32(data.SfxVolume))
	tdaudio.SetMusicVolume(float32(data.MusicVolume))
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

func (settings *Data) TextScale() float32 {
	if settings == nil || settings.WindowWidth > 800 {
		return 1.0
	}
	return 0.5
}

func UIWidth() float32 {
	return float32(Current.WindowWidth)
}

func UIHeight() float32 {
	return float32(Current.WindowHeight)
}
