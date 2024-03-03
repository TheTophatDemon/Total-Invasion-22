package world

import (
	"github.com/go-gl/mathgl/mgl32"
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
	owner      scene.Id[HasActor]
	sprite     scene.Id[*ui.Box]
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

func NewSickle(world *World, owner scene.Id[HasActor]) Weapon {
	return Weapon{
		owner: owner,
		onEquip: func(w *Weapon) {
			cache.GetSfx(SFX_SICKLE_THROW)
			cache.GetTexture(TEX_SICKLE_HUD)
		},
		onSelect: func(w *Weapon) {
			var (
				spriteBox *ui.Box
				err       error
			)
			w.sprite, spriteBox, err = world.UI.Boxes.New()
			if err != nil {
				return
			}
			spriteBox.
				SetSrc(math2.Rect{
					X: 256.0, Y: 0.0,
					Width: 256.0, Height: 192.0,
				}).
				SetDest(math2.Rect{
					X:     settings.UI_WIDTH - 256.0*2.0,
					Y:     settings.UI_HEIGHT - 192.0*2.0,
					Width: 256.0 * 2.0, Height: 192.0 * 2.0,
				}).
				SetTexture(cache.GetTexture(TEX_SICKLE_HUD)).
				SetColor(color.White)
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
					if ownerActor, ok := w.owner.Get(); ok {
						firePos := mgl32.TransformCoordinate(math2.Vec3Forward(), ownerActor.Body().Transform.Matrix())
						SpawnSickle(world.Projectiles, firePos, ownerActor.Body().Transform.Rotation(), w.owner.Handle)
					}
				}
			}
			if box, ok := w.sprite.Get(); ok {
				box.FlippedHorz = !box.FlippedHorz
			}
		},
	}
}
