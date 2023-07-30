package ecomps

import (
	"tophatdemon.com/total-invasion-ii/engine/ecs"
)

// Declare component storage
var AnimationPlayers *ecs.ComponentStorage[AnimationPlayer]
var Cameras *ecs.ComponentStorage[Camera]
var FirstPersonControllers *ecs.ComponentStorage[FirstPersonController]
var MeshRenders *ecs.ComponentStorage[MeshRender]
var Movements *ecs.ComponentStorage[Movement]
var Transforms *ecs.ComponentStorage[Transform]

// Register default components
func RegisterDefault(scene *ecs.Scene) {
	AnimationPlayers = ecs.RegisterComponent[AnimationPlayer](scene)
	Cameras = ecs.RegisterComponent[Camera](scene)
	FirstPersonControllers = ecs.RegisterComponent[FirstPersonController](scene)
	MeshRenders = ecs.RegisterComponent[MeshRender](scene)
	Movements = ecs.RegisterComponent[Movement](scene)
	Transforms = ecs.RegisterComponent[Transform](scene)
}
