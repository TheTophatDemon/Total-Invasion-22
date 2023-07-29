// Code generated DO NOT EDIT.

package ecomps

import (
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

// Declare component storage
var AnimationPlayerComps *scene.ComponentStorage[AnimationPlayer]
var CameraComps *scene.ComponentStorage[Camera]
var FirstPersonControllerComps *scene.ComponentStorage[FirstPersonController]
var MeshRenderComps *scene.ComponentStorage[MeshRender]
var MovementComps *scene.ComponentStorage[Movement]
var TransformComps *scene.ComponentStorage[Transform]

// Register default components
func RegisterAll(sc *scene.Scene) { 
    AnimationPlayerComps = scene.RegisterComponent[AnimationPlayer](sc)
    CameraComps = scene.RegisterComponent[Camera](sc)
    FirstPersonControllerComps = scene.RegisterComponent[FirstPersonController](sc)
    MeshRenderComps = scene.RegisterComponent[MeshRender](sc)
    MovementComps = scene.RegisterComponent[Movement](sc)
    TransformComps = scene.RegisterComponent[Transform](sc)
}