package settings

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

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
	// Menu actions
	ActionMenuDown      input.Action = "menuDown"
	ActionMenuUp        input.Action = "menuUp"
	ActionMenuConfirm   input.Action = "menuConfirm"
	ActionMenuClick     input.Action = "menuClick"
	ActionMenuCancel    input.Action = "menuCancel"
	ActionMenuIncrement input.Action = "menuIncrement"
	ActionMenuDecrement input.Action = "menuDecrement"

	// In game actions
	ActionForward  input.Action = "moveForward"
	ActionBack     input.Action = "moveBack"
	ActionLeft     input.Action = "moveLeft"
	ActionRight    input.Action = "moveRight"
	ActionSlow     input.Action = "slowDown"
	ActionLookHorz input.Action = "lookHorz"
	ActionLookVert input.Action = "lookVert"
	ActionFire     input.Action = "fire"

	// Weapon selection actions should be in the same order as
	// the WeaponType constants.
	ActionSickle        input.Action = "sickle"
	ActionChicken       input.Action = "chicken"
	ActionGrenade       input.Action = "grenade"
	ActionParusu        input.Action = "parusu"
	ActionDblGrenade    input.Action = "dblGrenade"
	ActionSign          input.Action = "sign"
	ActionAirhorn       input.Action = "airhorn"
	ActionDefenestrator input.Action = "defenestrator"
	ActionCluckster     input.Action = "cluckster"

	ActionUse         input.Action = "use"
	ActionWeaponWheel input.Action = "weaponWheel"

	// Cheat codes
	ActionNoclip       input.Action = "noclip"
	ActionGodMode      input.Action = "godMode"
	ActionMarySue      input.Action = "marySue"
	ActionDie          input.Action = "die"
	ActionKillEnemies  input.Action = "killEnemies"
	ActionCastBlessing input.Action = "castBlessing"
	ActionLaunchEditor input.Action = "launchEditor"
	ActionSpawnChicken input.Action = "spawnChicken"
	ActionLevelSelect  input.Action = "levelSelect"
)

const settingsFilePath = "game_settings.toml"

type Data struct {
	WindowWidth, WindowHeight uint16
	Fullscreen, Vsync         bool
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
		Fullscreen:       false,
		Vsync:            true,
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
		failure.LogErrWithLocation("failed to retrieve strings in %v for key %v: %v", Current.Locale, key, err)
		return "ERROR"
	}
	localizedText, ok := (*trans)[key]
	if !ok {
		// Fall back to English
		trans, err = cache.GetTranslation(fmt.Sprintf("assets/translations/strings_%v.toml", LocaleEnglish))
		if err != nil {
			failure.LogErrWithLocation("failed to retrieve English fallback for localization key %v: %v", key, err)
			return "ERROR"
		}
		localizedText = (*trans)[key]
	}

	// Parse as a text template in order to substitute control names and such.
	templ, err := template.New(key).
		Funcs(template.FuncMap{
			"binding": func(actionName string) string {
				bindings, ok := input.ActionBindings(input.Action(actionName))
				if !ok {
					return "UNKNOWN"
				}
				var name strings.Builder
				count := 0
				for _, bind := range bindings {
					if bind != nil {
						if count > 0 {
							name.WriteRune('/')
						}
						name.WriteString(bind.Name())
						count++
					}
				}

				return name.String()
			},
		}).
		Parse(localizedText)

	if err != nil {
		failure.LogErrWithLocation("failed to parse template for key %v in lang %v: %v", key, Current.Locale, err)
		return "ERROR"
	}

	var finalText strings.Builder
	err = templ.Execute(&finalText, nil)
	if err != nil {
		failure.LogErrWithLocation("failed to execute template for key %v in lang %v: %v", key, Current.Locale, err)
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
