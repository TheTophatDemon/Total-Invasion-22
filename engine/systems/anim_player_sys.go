package systems

import (
	"tophatdemon.com/total-invasion-ii/engine/comps"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
)

func UpdateAnimationPlayers(
	deltaTime float32,
	world *ecs.World,
	animPlayers *ecs.ComponentStorage[comps.AnimationPlayer]) {

	ents := world.Query(func(id ecs.EntID)bool{
		return animPlayers.Has(id)
	})

	for _, ent := range ents {
		animPlayer, _ := animPlayers.Get(ent)
		animPlayer.Update(deltaTime)
	}
}