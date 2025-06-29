package weapons

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

func newSickle() Weapon {
	return Weapon{
		name:          settings.Localize("sickle"),
		kind:          WeaponSickle,
		cooldown:      0.25,
		texturePath:   "assets/textures/ui/sickle_hud.png",
		swayExtents:   mgl32.Vec2{32.0, 16.0},
		swaySpeed:     mgl32.Vec2{0.75, 1.5},
		ammoType:      game.AMMO_TYPE_SICKLE,
		ammoCost:      1,
		spriteOffset:  mgl32.Vec2{settings.UIWidth() / 4.0, 0.0},
		updateFunc:    sickleUpdate,
		wheelColor:    color.FromBytes(138, 138, 138, 255),
		wheelIconPath: "assets/textures/sprites/sickle.png",
		noiseLevel:    10.0,
		isShooter:     true,
	}
}

func sickleUpdate(sickle *Weapon, deltaTime float32, ammo *game.Ammo) {
	throwAnim, _ := sickle.sprite.Texture.GetAnimation("fire")
	animPlayer := &sickle.sprite.AnimPlayer
	if sickle.CanFire(ammo) && animPlayer.CurrentAnimation().Name == throwAnim.Name {
		catchAnim, _ := sickle.sprite.Texture.GetAnimation("catch")
		animPlayer.ChangeAnimation(catchAnim)
		animPlayer.PlayFromStart()
		cache.GetSfx("assets/sounds/weapon/sickle_return.wav").Play()
	}
}
