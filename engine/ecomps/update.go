package ecomps

import "tophatdemon.com/total-invasion-ii/engine/scene"

// Update default components
func UpdateDefaultComps(sc *scene.Scene, ent scene.Entity, deltaTime float32) {
	animationPlayer, hasAnimationPlayer := AnimationPlayerComps.Get(ent)
	camera, hasCamera := CameraComps.Get(ent)
	firstPersonController, hasFirstPersonController := FirstPersonControllerComps.Get(ent)
	meshRender, hasMeshRender := MeshRenderComps.Get(ent)
	movement, hasMovement := MovementComps.Get(ent)
	transform, hasTransform := TransformComps.Get(ent)

	_, _ = meshRender, hasMeshRender
	_, _ = camera, hasCamera

	if hasMovement && hasTransform {
		movement.Update(transform, ent, deltaTime)
		if hasFirstPersonController {
			firstPersonController.Update(movement, ent, deltaTime)
		}
	}

	if hasAnimationPlayer {
		animationPlayer.Update(deltaTime)
	}
}
