package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/audio"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type WeaponSickle struct {
	weaponBase
	sfxCatch             *audio.Sfx
	hudTexture           *textures.Texture
	throwAnim, catchAnim textures.Animation
	thrownSickle         scene.Id[*Projectile]
}

var _ Weapon = (*WeaponSickle)(nil)

func NewSickle(world *World, owner scene.Id[HasActor]) WeaponSickle {
	sickle := WeaponSickle{
		weaponBase: weaponBase{
			owner: owner,
			world: world,
		},
	}

	var (
		err error
		ok  bool
	)

	sickle.sfxCatch, err = cache.GetSfx("assets/sounds/sickle_return.wav")
	if err != nil {
		log.Println(err)
	}

	sickle.hudTexture = cache.GetTexture("assets/textures/ui/sickle_hud.png")

	sickle.throwAnim, ok = sickle.hudTexture.GetAnimation("throw")
	if !ok {
		log.Println("sickle throw anim not found")
	}
	sickle.catchAnim, ok = sickle.hudTexture.GetAnimation("catch")
	if !ok {
		log.Println("sickle catch anim not found")
	}

	return sickle
}

func (sickle *WeaponSickle) Order() int {
	return WEAPON_ORDER_SICKLE
}

func (sickle *WeaponSickle) Equip() {
	sickle.weaponBase.Equip()
}

func (sickle *WeaponSickle) Select() {
	var (
		spriteBox *ui.Box
		err       error
	)
	sickle.sprite, spriteBox, err = sickle.world.UI.Boxes.New()
	if err != nil {
		log.Println(err)
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
		SetTexture(sickle.hudTexture).
		SetColor(color.White)

	spriteBox.AnimPlayer.ChangeAnimation(sickle.throwAnim)
}

func (sickle *WeaponSickle) Deselect() {
	sickle.sprite.Remove()
}

func (sickle *WeaponSickle) Update(deltaTime float32) {
	var sprite *ui.Box
	var ok bool
	if sprite, ok = sickle.sprite.Get(); !ok {
		return
	}

	sprite.AnimPlayer.Update(deltaTime)

	if !sickle.thrownSickle.Exists() && sprite.AnimPlayer.CurrentAnimation().Name == sickle.throwAnim.Name {
		sprite.AnimPlayer.ChangeAnimation(sickle.catchAnim)
		sprite.AnimPlayer.PlayFromStart()
		sickle.sfxCatch.Play()
	}
}

func (sickle *WeaponSickle) Fire() {
	if !sickle.thrownSickle.Exists() {
		if ownerActor, ok := sickle.owner.Get(); ok {
			firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, 0.0, -0.5}, ownerActor.Body().Transform.Matrix())
			sickle.thrownSickle, _, _ = SpawnSickle(sickle.world, &sickle.world.Projectiles, firePos, ownerActor.Body().Transform.Rotation(), sickle.owner.Handle)
			if box, ok := sickle.sprite.Get(); ok {
				box.AnimPlayer.ChangeAnimation(sickle.throwAnim)
				box.AnimPlayer.PlayFromStart()
			}
		}
	}
}
