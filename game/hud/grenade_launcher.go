package hud

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type GrenadeLauncher struct {
	weaponBase
}

func (grenadeLauncher *GrenadeLauncher) Init() {
	grenadeLauncher.weaponBase = weaponBase{
		cooldown:      1.0,
		spriteTexture: cache.GetTexture("assets/textures/ui/grenade_launcher_hud.png"),
		swayExtents:   mgl32.Vec2{16.0, 8.0},
		swaySpeed:     mgl32.Vec2{0.75, 1.25},
		ammoType:      game.AMMO_TYPE_GRENADE,
		ammoCost:      1,
	}

	idleAnim, ok := grenadeLauncher.spriteTexture.GetAnimation("idle")
	if !ok {
		log.Println("grenade launcher idle anim not found")
	}
	grenadeLauncher.defaultAnimation = idleAnim

	grenadeLauncher.spriteSize = mgl32.Vec2{
		idleAnim.Frames[0].Rect.Width * SpriteScale(),
		idleAnim.Frames[0].Rect.Height * SpriteScale(),
	}
	grenadeLauncher.spriteEndPos = mgl32.Vec2{
		settings.UIWidth()/2 - grenadeLauncher.spriteSize.X()/2.0,
		settings.UIHeight() - grenadeLauncher.spriteSize.Y(),
	}
	grenadeLauncher.spriteStartPos = grenadeLauncher.spriteEndPos.Add(mgl32.Vec2{0.0, grenadeLauncher.spriteSize.Y()})
}

func (grenadeLauncher *GrenadeLauncher) Order() WeaponIndex {
	return WEAPON_ORDER_GRENADE
}

func (grenadeLauncher *GrenadeLauncher) NoiseLevel() float32 {
	return 50.0
}

func (grenadeLauncher *GrenadeLauncher) Update(deltaTime float32, swayAmount float32, ammo *game.Ammo) {
	grenadeLauncher.weaponBase.Update(deltaTime, swayAmount, ammo)

	if grenadeLauncher.CanFire(ammo) || ammo[grenadeLauncher.AmmoType()] == 0 {
		grenadeLauncher.sprite.AnimPlayer.ChangeAnimation(grenadeLauncher.defaultAnimation)
	}
}

func (grenadeLauncher *GrenadeLauncher) Fire(ammo *game.Ammo) {
	grenadeLauncher.weaponBase.Fire(ammo)
	animPlayer := &grenadeLauncher.sprite.AnimPlayer
	fireAnim, ok := grenadeLauncher.spriteTexture.GetAnimation("fire")
	if !ok {
		log.Println("grenade launcher fire anim not found")
		return
	}
	if animPlayer.CurrentAnimation().Name != fireAnim.Name {
		animPlayer.ChangeAnimation(fireAnim)
		animPlayer.PlayFromStart()
	}
}

func (grenadeLauncher *GrenadeLauncher) IsShooter() bool {
	return true
}

func (grenadeLauncher *GrenadeLauncher) WheelColor() color.Color {
	return color.FromBytes(0, 170, 0, 255)
}

func (grenadeLauncher *GrenadeLauncher) WheelIconPath() string {
	return "assets/textures/sprites/grenade_launcher.png"
}
