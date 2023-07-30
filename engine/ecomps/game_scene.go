package ecomps

import (
	"tophatdemon.com/total-invasion-ii/engine/ecs"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type GameScene struct {
	ecs.Scene
	AnimationPlayers       *ecs.ComponentStorage[AnimationPlayer]
	Cameras                *ecs.ComponentStorage[Camera]
	FirstPersonControllers *ecs.ComponentStorage[FirstPersonController]
	MeshRenders            *ecs.ComponentStorage[MeshRender]
	Movements              *ecs.ComponentStorage[Movement]
	Transforms             *ecs.ComponentStorage[Transform]
}

// Register default components
func NewGameScene(maxEnts uint) GameScene {
	return GameScene{
		ecs.NewScene(maxEnts),
		ecs.NewStorage[AnimationPlayer](maxEnts),
		ecs.NewStorage[Camera](maxEnts),
		ecs.NewStorage[FirstPersonController](maxEnts),
		ecs.NewStorage[MeshRender](maxEnts),
		ecs.NewStorage[Movement](maxEnts),
		ecs.NewStorage[Transform](maxEnts),
	}
}

// Render default components
func (scene *GameScene) Render(ent ecs.Entity, context *render.Context) {
	transform, _ := scene.Transforms.Get(ent)
	animPlayer, _ := scene.AnimationPlayers.Get(ent)
	meshRender, hasMeshRender := scene.MeshRenders.Get(ent)
	if hasMeshRender {
		meshRender.Render(transform, animPlayer, ent, context)
	}
}

// Update default components
func (scene *GameScene) Update(ent ecs.Entity, deltaTime float32) {
	animationPlayer, hasAnimationPlayer := scene.AnimationPlayers.Get(ent)
	firstPersonController, hasFirstPersonController := scene.FirstPersonControllers.Get(ent)
	movement, hasMovement := scene.Movements.Get(ent)
	transform, hasTransform := scene.Transforms.Get(ent)

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
