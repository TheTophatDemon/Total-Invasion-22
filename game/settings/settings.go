package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/locales"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
)

const (
	SETTINGS_FILE_PATH = "game_settings.json"
)

const (
	ACTION_FORWARD input.Action = iota
	ACTION_BACK
	ACTION_LEFT
	ACTION_RIGHT
	ACTION_SLOW
	ACTION_LOOK_HORZ
	ACTION_LOOK_VERT
	ACTION_TRAP_MOUSE
	ACTION_FIRE
	ACTION_SICKLE
	ACTION_CHICKEN
	ACTION_USE
	ACTION_NOCLIP
	ACTION_GODMODE
	ACTION_COUNT
)

var actionNames = [ACTION_COUNT]string{
	ACTION_FORWARD:    "Move Forward",
	ACTION_BACK:       "Move Back",
	ACTION_LEFT:       "Strafe Left",
	ACTION_RIGHT:      "Strafe Right",
	ACTION_SLOW:       "Slow",
	ACTION_LOOK_HORZ:  "Look Horizontally",
	ACTION_LOOK_VERT:  "Look Vertically",
	ACTION_TRAP_MOUSE: "Trap Mouse",
	ACTION_FIRE:       "Fire",
	ACTION_SICKLE:     "Select Sickle",
	ACTION_CHICKEN:    "Select Chicken Cannon",
	ACTION_USE:        "Use",
	ACTION_NOCLIP:     "Noclip",
	ACTION_GODMODE:    "God Mode",
}

type Data struct {
	WindowWidth, WindowHeight uint16
	MouseSensitivity          float32
	TextShadowColor           color.Color
	SfxVolume, MusicVolume    float32
	Locale                    string
	Fov                       float32 // Measured in degrees
	Debug                     struct {
		StartMap string
	}
	DifficultyIndex int
}

func (data *Data) WindowAspectRatio() float32 {
	return float32(data.WindowWidth) / float32(data.WindowHeight)
}

var Default, Current Data

func init() {
	Default = Data{
		WindowWidth: 1280, WindowHeight: 720,
		MouseSensitivity: 0.005,
		TextShadowColor:  color.Color{R: 0.0, G: 0.0, B: 0.0, A: 0.5},
		SfxVolume:        1.0, MusicVolume: 1.0,
		Locale:          locales.ENGLISH,
		Fov:             70.0,
		DifficultyIndex: len(Difficulties) - 1,
	}
	Current = Default
}

func ActionName(action input.Action) string {
	if action > ACTION_COUNT {
		return ""
	}
	return actionNames[action]
}

func LoadOrInit() {
	settingsFile, err := os.Open(SETTINGS_FILE_PATH)
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

		err = json.Unmarshal(fileBytes, &Current)
		if err != nil {
			failure.LogErrWithLocation("Could not unmarshal settings file; %v", err)
			Current = Default
			return
		}
	}
}

func Save() {
	settingsFile, err := os.Create(SETTINGS_FILE_PATH)
	if err != nil {
		failure.LogErrWithLocation("Could not open settings file for writing; %v", err)
	}
	defer settingsFile.Close()

	settingsBytes, err := json.Marshal(Current)
	if err != nil {
		failure.LogErrWithLocation("Could not marshal settings; %v", err)
		return
	}

	var formattedBuffer bytes.Buffer
	json.Indent(&formattedBuffer, settingsBytes, "", "\t")
	_, err = formattedBuffer.WriteTo(settingsFile)

	if err != nil {
		failure.LogErrWithLocation("Could not write settings to file; %v", err)
		return
	}
}

func Localize(key string) string {
	trans, err := cache.GetTranslation(fmt.Sprintf("assets/translations/%v.json", strings.ToLower(Current.Locale)))
	if err != nil {
		return "ERROR"
	}
	localizedText, ok := (*trans)[key]
	if !ok {
		// Fall back to English
		trans, err = cache.GetTranslation(fmt.Sprintf("assets/translations/%v.json", locales.ENGLISH))
		if err != nil {
			return "ERROR"
		}
		localizedText = (*trans)[key]
	}
	return localizedText
}

func UIScale() float32 {
	return float32(Current.WindowHeight) / 480
}

func UIWidth() float32 {
	return float32(Current.WindowWidth)
}

func UIHeight() float32 {
	return float32(Current.WindowHeight)
}
