package world

import (
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

func (w *World) BodiesIter() func() (comps.HasBody, scene.Handle) {
	playerIter := w.Players.Iter()
	enemiesIter := w.Enemies.Iter()
	wallsIter := w.Walls.Iter()
	propsIter := w.Props.Iter()
	projsIter := w.Projectiles.Iter()
	return func() (comps.HasBody, scene.Handle) {
		if player, id := playerIter(); player != nil {
			return player, id
		}
		if enemy, id := enemiesIter(); enemy != nil {
			return enemy, id
		}
		if wall, id := wallsIter(); wall != nil {
			return wall, id
		}
		if prop, id := propsIter(); prop != nil {
			return prop, id
		}
		if proj, id := projsIter(); proj != nil {
			return proj, id
		}
		return nil, scene.Handle{}
	}
}

func (w *World) ActorsIter() func() (HasActor, scene.Handle) {
	playerIter := w.Players.Iter()
	enemiesIter := w.Enemies.Iter()
	return func() (HasActor, scene.Handle) {
		if player, id := playerIter(); player != nil {
			return player, id
		}
		if enemy, id := enemiesIter(); enemy != nil {
			return enemy, id
		}
		return nil, scene.Handle{}
	}
}

func (w *World) LinkablesIter(linkNumber int) func() (Linkable, scene.Handle) {
	triggerIter := w.Triggers.Iter()
	return func() (Linkable, scene.Handle) {
		if trigger, id := triggerIter(); trigger != nil {
			return trigger, id
		}
		return nil, scene.Handle{}
	}
}
