package ents

import (
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

const (
	WEAPON_ORDER_NONE int = iota
	WEAPON_ORDER_SICKLE
	WEAPON_ORDER_CHICKEN
	WEAPON_ORDER_GRENADE
	WEAPON_ORDER_DBL_GRENADE
	WEAPON_ORDER_SIGN
	WEAPON_ORDER_AIRHORN
	WEAPON_ORDER_MAX
)

type weaponBase struct {
	actor  HasActor
	sprite scene.Id[ui.Box]
	world  WorldOps
}
