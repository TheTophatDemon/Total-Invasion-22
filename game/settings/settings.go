package settings

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/audio"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/input"
)

const (
	UI_HEIGHT = 480
	UI_WIDTH  = 800
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
	ACTION_MUTE_MUS
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
	ACTION_MUTE_MUS:   "Mute Music",
}

type Data struct {
	WindowWidth, WindowHeight uint16
	MouseSensitivity          float32
	sfxVolume, musicVolume    float32
}

func (data *Data) WindowAspectRatio() float32 {
	return float32(data.WindowWidth) / float32(data.WindowHeight)
}

func (data *Data) SfxVolume() float32 {
	return data.sfxVolume
}

func (data *Data) MusicVolume() float32 {
	return data.musicVolume
}

func UpdateMusicVolume(target float32) {
	Current.musicVolume = target
	// Set the volume of all currently playing songs
	cache.IterateSongs()(func(path string, song *audio.Song) bool {
		if song.IsPlaying() {
			song.SetVolume(Current.musicVolume)
		}
		return true
	})
}

var Default, Current Data

func init() {
	Default = Data{
		WindowWidth: 1280, WindowHeight: 720,
		MouseSensitivity: 0.005,
		sfxVolume:        1.0, musicVolume: 1.0,
	}
	Current = Default
}

func ActionName(action input.Action) string {
	if action > ACTION_COUNT {
		return ""
	}
	return actionNames[action]
}
