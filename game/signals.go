package game

import (
	jsonv1 "encoding/json"
	"encoding/json/v2"
	"time"

	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

type (
	// Acts as both a message to change the map and to load saved entity data.
	MapChangeSignal struct {
		// File path to the map to change to, including the extension, relative to the game directory.
		MapPath     string
		MapTitleKey string
		// This will override the state of the player after she is loaded, superceding what is in the SavedEnts array.
		PlayerEnt *EntDef
		// State of entities from a loaded save file, if applicable.
		SavedEnts []EntDef
		// Time when the save was made.
		Timestamp              time.Time
		KillCount, SecretCount uint
		TimeSoFar              time.Duration
		AfterCheckpoint        bool
		DifficultyIndex        int

		// Indicates that the world should make a new autosave after it loads
		SaveAfterLoad bool `json:"-"`
	}
	ResumeGameSignal   struct{}
	ChangeScreenSignal struct {
		Screen ui.Screen
	}
	TeleportationSignal struct{}
	SaveSignal          struct {
		Number          int
		AfterCheckpoint bool
		WithData        *MapChangeSignal
	}
	LoadSignal struct {
		Number int
		// The path to the map file that will stop the loading process if it does not match.
		RequireMap string
	}
)

func (ss SaveSignal) IsTemporary() bool {
	return ss.Number == 0
}

func (mcs MapChangeSignal) IsNil() bool {
	return len(mcs.MapPath) == 0
}

// Returns options that should be used when parsing a save file as JSON
func SaveFileParseOptions() json.Options {
	return json.JoinOptions(te3.ParseOptions(), jsonv1.FormatDurationAsNano(true))
}
