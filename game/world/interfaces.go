package world

type (
	// Represents an entity that reacts to having the 'use' key pressed when the player is pointing at it.
	Usable interface {
		OnUse(p *Player)
	}

	Observer interface {
		ProcessSignal(Signal, any)
	}

	// Represents an entity that can be activated by another entity.
	Linkable interface {
		LinkNumber() int
	}
)
