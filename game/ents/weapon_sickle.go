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
	SICKLE_TEX_PATH       = "assets/textures/ui/sickle_hud.png"
	SICKLE_THROW_SFX_PATH = "assets/sounds/sickle.wav"
)

type Sickle struct {
	weaponBase
	throwingSound audio.PlayingId
}

var _ Weapon = (*Sickle)(nil)

func NewSickle(actor HasActor, world WorldOps) *Sickle {
	return &Sickle{
		weaponBase: weaponBase{
			actor:  actor,
			sprite: scene.Id[ui.Box]{},
			world:  world,
		},
		throwingSound: audio.PlayingId(0),
	}
}

func (s *Sickle) Order() int {
	return WEAPON_ORDER_SICKLE
}

func (s *Sickle) OnEquip() {
	cache.GetSfx(SICKLE_THROW_SFX_PATH)
}

func (s *Sickle) OnSelect() {
	s.sprite, _ = s.world.AddUiBox(ui.NewBox(
		math2.Rect{
			X: 0.0, Y: 0.0,
			Width: 256.0, Height: 192.0,
		}, math2.Rect{
			X:     settings.WINDOW_WIDTH - 256.0,
			Y:     settings.WINDOW_HEIGHT - 192.0,
			Width: 256.0, Height: 192.0,
		},
		cache.GetTexture(SICKLE_TEX_PATH),
		color.White))
}

func (s *Sickle) OnFire() {
	if sickleSfx, err := cache.GetSfx(SICKLE_THROW_SFX_PATH); err == nil {
		if s.throwingSound != 0 {
			sickleSfx.Stop(s.throwingSound)
			s.throwingSound = audio.PlayingId(0)
		} else {
			s.throwingSound = sickleSfx.Play()
		}
	}
}

func (s *Sickle) OnDeselect() {
	s.sprite.Remove()
}

func (s *Sickle) Update(deltaTime float32) {

}
