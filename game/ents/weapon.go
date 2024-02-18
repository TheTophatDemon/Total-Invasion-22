package ents

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/audio"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
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

const (
	SFX_SICKLE_THROW = "assets/sounds/sickle.wav"
	TEX_SICKLE_HUD   = "assets/textures/ui/sickle_hud.png"
)

type Weapon struct {
	sprite     scene.Id[ui.Box]
	equipped   bool
	onEquip    func(w *Weapon)
	onSelect   func(w *Weapon)
	onDeselect func(w *Weapon)
	onFire     func(w *Weapon)
	fireSound  audio.PlayingId
}

func (w *Weapon) Fire() {
	if w.onFire != nil {
		w.onFire(w)
	}
}

func (w *Weapon) Select() {
	if w.onSelect != nil {
		w.onSelect(w)
	}
}

func (w *Weapon) Deselect() {
	if w.onDeselect != nil {
		w.onDeselect(w)
	}
}

func (w *Weapon) Equip() {
	if w.onEquip != nil {
		w.onEquip(w)
	}
	w.equipped = true
}

func NewSickle(world WorldOps) Weapon {
	return Weapon{
		onEquip: func(w *Weapon) {
			cache.GetSfx(SFX_SICKLE_THROW)
			cache.GetTexture(TEX_SICKLE_HUD)
		},
		onSelect: func(w *Weapon) {
			w.sprite, _ = world.AddUiBox(ui.NewBox(
				math2.Rect{
					X: 0.0, Y: 0.0,
					Width: 256.0, Height: 192.0,
				}, math2.Rect{
					X:     settings.WINDOW_WIDTH - 256.0,
					Y:     settings.WINDOW_HEIGHT - 192.0,
					Width: 256.0, Height: 192.0,
				},
				cache.GetTexture(TEX_SICKLE_HUD),
				color.White))
		},
		onDeselect: func(w *Weapon) {
			w.sprite.Remove()
		},
		onFire: func(w *Weapon) {
			if sickleSfx, err := cache.GetSfx(SFX_SICKLE_THROW); err == nil {
				if w.fireSound != 0 {
					sickleSfx.Stop(w.fireSound)
					w.fireSound = audio.PlayingId(0)
				} else {
					w.fireSound = sickleSfx.Play()
				}
			}
		},
	}
}
