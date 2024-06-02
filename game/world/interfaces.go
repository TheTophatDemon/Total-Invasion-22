package world

import "tophatdemon.com/total-invasion-ii/engine/scene/comps"

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

	HasActor interface {
		comps.HasBody
		Observer
		Actor() *Actor
	}

	Weapon interface {
		Order() WeaponIndex
		Equip()
		IsEquipped() bool
		Select()
		Deselect()
		IsSelected() bool
		Fire()
		CanFire() bool
		Update(deltaTime float32, swayAmount float32)
	}
)
