package ecomps

import "tophatdemon.com/total-invasion-ii/engine/ecs"

// Update default components
func UpdateDefaultComps(scene *ecs.Scene, ent ecs.Entity, deltaTime float32) {
	animationPlayer, hasAnimationPlayer := AnimationPlayers.Get(ent)
	camera, hasCamera := Cameras.Get(ent)
	firstPersonController, hasFirstPersonController := FirstPersonControllers.Get(ent)
	meshRender, hasMeshRender := MeshRenders.Get(ent)
	movement, hasMovement := Movements.Get(ent)
	transform, hasTransform := Transforms.Get(ent)

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
