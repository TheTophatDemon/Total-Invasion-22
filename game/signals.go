package game

import (
	"time"

	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

type (
	// Acts as both a message to change the map and to load saved entity data.
	MapChangeSignal struct {
		// File path to the map to change to, including the extension, relative to the game directory.
		MapPath string
		// This will override the state of the player after she is loaded, superceding what is in the Ents array.
		PlayerEnt *te3.Ent
		// State of entities from a loaded save file, if applicable.
		Ents []te3.Ent
		// Time when the save was made.
		Timestamp              time.Time
		KillCount, SecretCount uint
		TimeSoFar              time.Duration
	}
	ResumeGameSignal   struct{}
	ChangeScreenSignal struct {
		Screen ui.Screen
	}
	TeleportationSignal struct{}
	SaveSignal          struct {
		Number int
	}
)

func (ss SaveSignal) IsTemporary() bool {
	return ss.Number == 0
}
