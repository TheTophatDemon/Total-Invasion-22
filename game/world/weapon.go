package world

import (
	"tophatdemon.com/total-invasion-ii/engine/audio"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

const (
	WEAPON_ORDER_NONE int = iota - 1
	WEAPON_ORDER_SICKLE
	WEAPON_ORDER_CHICKEN
	WEAPON_ORDER_GRENADE
	WEAPON_ORDER_PARUSU
	WEAPON_ORDER_DBL_GRENADE
	WEAPON_ORDER_SIGN
	WEAPON_ORDER_AIRHORN
	WEAPON_ORDER_MAX
)

type weaponBase struct {
	owner     scene.Id[HasActor]
	sprite    scene.Id[*ui.Box]
	equipped  bool
	fireVoice audio.VoiceId
	world     *World
}

func (wb *weaponBase) Equip() {
	wb.equipped = true
}

func (wb *weaponBase) Equipped() bool {
	return wb.equipped
}
