package world

import (
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type (
	// Represents an entity that reacts to having the 'use' key pressed when the player is pointing at it.
	Usable interface {
		OnUse(p *Player)
	}

	Damageable interface {
		OnDamage(sourceEntity any, amount float32) bool
	}

	// Represents an entity that can be activated by another entity.
	Linkable interface {
		scene.HasHandle
		LinkNumber() int
		OnLinkActivate(source Linkable)
	}

	HasActor interface {
		comps.HasBody
		Damageable
		engine.Observer
		Actor() *Actor
	}
)
