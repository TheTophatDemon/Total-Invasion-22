package game

import (
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/ents"
)

func (w *World) BodiesIter() func() (comps.HasBody, scene.Handle) {
	playerIter := w.Players.Iter()
	enemiesIter := w.Enemies.Iter()
	wallsIter := w.Walls.Iter()
	propsIter := w.Props.Iter()
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
		return nil, nil
	}
}

func (w *World) ActorsIter() func() (ents.HasActor, scene.Handle) {
	playerIter := w.Players.Iter()
	enemiesIter := w.Enemies.Iter()
	return func() (ents.HasActor, scene.Handle) {
		if player, id := playerIter(); player != nil {
			return player, id
		}
		if enemy, id := enemiesIter(); enemy != nil {
			return enemy, id
		}
		return nil, nil
	}
}

func (w *World) LinkablesIter(linkNumber int) func() (ents.Linkable, scene.Handle) {
	triggerIter := w.Triggers.Iter()
	return func() (ents.Linkable, scene.Handle) {
		if trigger, id := triggerIter(); trigger != nil {
			return trigger, id
		}
		return nil, nil
	}
}
