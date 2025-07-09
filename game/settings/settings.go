package settings

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/input"
)

const (
	LocaleEnglish = "en"
	LocaleRussian = "ru"
)

const (
	ActionForward input.Action = iota
	ActionBack
	ActionLeft
	ActionRight
	ActionSlow
	ActionLookHorz
	ActionLookVert
	ActionTrapMouse
	ActionFire

	// Weapon selection actions should be in the same order as
	// the WeaponType constants.
	ActionSickle
	ActionChicken
	ActionGrenade
	ActionParusu
	ActionDblGrenade
	ActionSign
	ActionAirhorn
	ActionDefenestrator
	ActionCluckster

	ActionUse
	ActionWeaponWheel
	ActionNoclip
	ActionGodMode
	ActionMarySue
	ActionDie
	ActionKillEnemies
	ActionCastBlessing
	ActionCount
)

const settingsFilePath = "game_settings.toml"

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

var Default, Current Data

func init() {
	Default = Data{
		WindowWidth: 1280, WindowHeight: 720,
		MouseSensitivity: 0.005,
		TextShadowColor:  color.Color{R: 0.0, G: 0.0, B: 0.0, A: 0.5},
		SfxVolume:        1.0, MusicVolume: 1.0,
		Locale:          LocaleEnglish,
		Fov:             70.0,
		DifficultyIndex: len(Difficulties) - 1,
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
	trans, err := cache.GetTranslation(fmt.Sprintf("assets/translations/strings_%v.toml", strings.ToLower(Current.Locale)))
	if err != nil {
		return "ERROR"
	}
	localizedText, ok := (*trans)[key]
	if !ok {
		// Fall back to English
		trans, err = cache.GetTranslation(fmt.Sprintf("assets/translations/strings_%v.toml", LocaleEnglish))
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

// Returns the size the sprites on the HUD should be scaled to.
func SpriteScale() float32 {
	return UIScale() * 2.0
}

func UIWidth() float32 {
	return float32(Current.WindowWidth)
}

func UIHeight() float32 {
	return float32(Current.WindowHeight)
}
