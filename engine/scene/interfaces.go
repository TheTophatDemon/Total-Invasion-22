package scene

import "tophatdemon.com/total-invasion-ii/engine/render"

type (
	// Type-agnostic abstraction of Storage
	StorageOps interface {
		GetUntyped(Handle) (any, bool)
		Has(Handle) bool
		Remove(Handle)
		Update(deltaTime float32)
		Render(renderContext *render.Context)
		TearDown() // Called to release external resources held by storage items.
	}

	HasHandle interface {
		Handle() Handle
	}
)
