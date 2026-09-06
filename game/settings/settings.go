package settings

import (
	"errors"
	"io"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/go-gl/glfw/v3.3/glfw"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
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
		ActionLookLeft:      Action{input.NewMouseMovementBinding(input.MouseAxisNegX), input.NewKeyBinding(glfw.KeyLeft)},
		ActionFastLookLeft:  Action{nil, nil},
		ActionLookRight:     Action{input.NewMouseMovementBinding(input.MouseAxisPosX), input.NewKeyBinding(glfw.KeyRight)},
		ActionFastLookRight: Action{nil, nil},
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
