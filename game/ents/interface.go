package ents

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

type (
	// In order to prevent a circular dependency of packages, entities interact with the World through this interface.
	// As a bonus, this prevents entities from doing things with the world that they shouldn't, like running a full update.
	WorldOps interface {
		ShowMessage(text string, duration float32, priority int, colr color.Color)
		FlashScreen(color color.Color, fadeSpeed float32)
		Raycast(rayOrigin, rayDir mgl32.Vec3, includeBodies bool, maxDist float32, excludeBody comps.HasBody) (collision.RaycastResult, scene.Handle)
		BodiesInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception comps.HasBody) []scene.Handle
		ActorsInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception HasActor) []scene.Handle
		LinkablesIter(linkNumber int) func() (Linkable, scene.Handle)
		AddUiBox(box ui.Box) (scene.Id[ui.Box], error)
	}

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
