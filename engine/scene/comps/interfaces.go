package comps

import "tophatdemon.com/total-invasion-ii/engine/scene"

type (
	HasBody interface {
		Body() *Body
	}

	BodyIter interface {
		Next() (HasBody, scene.Handle)
	}
)
