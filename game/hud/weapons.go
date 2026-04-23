package hud

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Weapons struct {
	hud                        *Hud
	weapons                    [game.WeaponCount]Weapon
	weaponWheel                WeaponWheel
	selectedWeapon, nextWeapon game.WeaponType
}

func (weapons *Weapons) init(hud *Hud) {
	*weapons = Weapons{
		hud: hud,
		weapons: [game.WeaponCount]Weapon{
			// Sickle
			game.WeaponSickle: {
				name:            "sickle",
				kind:            game.WeaponSickle,
				cooldown:        0.25,
				texturePath:     "assets/textures/ui/sickle_hud.png",
				initialAnimName: fireAnimName,
				swayExtents:     mgl32.Vec2{32.0, 16.0},
				swaySpeed:       mgl32.Vec2{0.75, 1.5},
				ammoType:        game.AmmoTypeSickle,
				ammoCost:        1,
				spriteOffset:    mgl32.Vec2{settings.UIWidth() / 4.0, 0.0},
				updateFunc: func(sickle *Weapon, deltaTime float32, ammo game.Ammo) {
					animPlayer := &sickle.sprite.AnimPlayer
					if sickle.canFire(ammo) && animPlayer.CurrentAnimation().Name == fireAnimName {
						catchAnim, _ := sickle.sprite.BgTexture.GetAnimation("catch")
						animPlayer.ChangeAnimation(catchAnim)
						animPlayer.PlayFromStart()
						cache.GetSfx("assets/sounds/weapon/sickle_return.wav").Play()
					}
				},
				wheelColor:    color.FromBytes(138, 138, 138, 255),
				wheelIconPath: "assets/textures/sprites/sickle.png",
				noiseLevel:    10.0,
				isShooter:     true,
			},
			// Chicken Cannon
			game.WeaponChicken: {
				name:          "chickenCannon",
				kind:          game.WeaponChicken,
				cooldown:      0.15,
				texturePath:   "assets/textures/ui/chicken_cannon_hud.png",
				swayExtents:   mgl32.Vec2{16.0, 8.0},
				swaySpeed:     mgl32.Vec2{0.5, 1.0},
				ammoType:      game.AmmoTypeEgg,
				ammoCost:      1,
				wheelColor:    color.FromBytes(0, 0, 255, 255),
				wheelIconPath: "assets/textures/sprites/chicken_cannon.png",
				noiseLevel:    20.0,
				isShooter:     true,
			},
			// Grenade launcher
			game.WeaponGrenade: {
				name:          "grenadeLauncher",
				kind:          game.WeaponGrenade,
				cooldown:      1.0,
				texturePath:   "assets/textures/ui/grenade_launcher_hud.png",
				swayExtents:   mgl32.Vec2{16.0, 8.0},
				swaySpeed:     mgl32.Vec2{0.75, 1.25},
				ammoType:      game.AmmoTypeGrenade,
				ammoCost:      1,
				wheelColor:    color.FromBytes(0, 170, 0, 255),
				wheelIconPath: "assets/textures/sprites/grenade_launcher.png",
				noiseLevel:    50.0,
				isShooter:     true,
			},
			// Parusu
			game.WeaponParusu: {
				name:          "parusu",
				kind:          game.WeaponParusu,
				cooldown:      0.075,
				texturePath:   "assets/textures/ui/parusu_hud.png",
				swayExtents:   mgl32.Vec2{16.0, 8.0},
				swaySpeed:     mgl32.Vec2{0.5, 1.0},
				ammoType:      game.AmmoTypePlasma,
				ammoCost:      1,
				wheelColor:    color.FromBytes(0, 255, 130, 255),
				wheelIconPath: "assets/textures/sprites/parusu.png",
				noiseLevel:    40.0,
				isShooter:     true,
				updateFunc: func(parusu *Weapon, deltaTime float32, ammo game.Ammo) {
					animPlayer := &parusu.sprite.AnimPlayer
					if animPlayer.CurrentAnimation().Name == fireAnimName && animPlayer.IsAtEnd() {
						idleAnim, _ := parusu.sprite.BgTexture.GetAnimation(idleAnimName)
						animPlayer.PlayNewAnim(idleAnim)
					}
				},
			},
			// Double grenade launcher (NOT IMPLEMENTED)
			game.WeaponDblGrenade: {
				name:          "doubleGrenadeLauncher",
				kind:          game.WeaponDblGrenade,
				wheelColor:    color.FromBytes(255, 130, 0, 255),
				wheelIconPath: "assets/textures/sprites/double_grenade_launcher.png",
			},
			// Sign of Madness (NOT IMPLEMENTED)
			game.WeaponSign: {
				name:          "signOfMadness",
				kind:          game.WeaponSign,
				wheelColor:    color.FromBytes(170, 0, 0, 255),
				wheelIconPath: "assets/textures/sprites/sign_of_madness.png",
			},
			// Airhorn
			game.WeaponAirhorn: {
				name:          "airhorn",
				kind:          game.WeaponAirhorn,
				cooldown:      0.0,
				texturePath:   "assets/textures/ui/airhorn_hud.png",
				swayExtents:   mgl32.Vec2{32.0, 16.0},
				swaySpeed:     mgl32.Vec2{0.75, 1.5},
				ammoType:      game.AmmoTypeNone,
				spriteOffset:  mgl32.Vec2{settings.UIWidth() / 6.0, 0.0},
				noiseLevel:    100.0,
				wheelColor:    color.FromBytes(255, 0, 0, 255),
				wheelIconPath: "assets/textures/sprites/airhorn.png",
				fireSoundPath: "assets/sounds/weapon/airhorn.wav",
				updateFunc: func(airhorn *Weapon, deltaTime float32, ammo game.Ammo) {
					if !airhorn.heldDown {
						idleAnim, _ := airhorn.sprite.BgTexture.GetAnimation(idleAnimName)
						airhorn.sprite.AnimPlayer.PlayNewAnim(idleAnim)
						if airhorn.voice.IsPlaying() && airhorn.voice.GetTime() < 800 {
							airhorn.voice.Seek(800)
						}
					}
				},
			},
			// Defenestrator (NOT IMPLEMENTED)
			game.WeaponDefenestrator: {
				name:       "defenestrator",
				kind:       game.WeaponDefenestrator,
				ammoType:   game.AmmoTypePlasma,
				wheelColor: color.FromBytes(32, 32, 32, 255),
			},
			// Cluckster Bomb (NOT IMPLEMENTED)
			game.WeaponCluckster: {
				name:       "clucksterBomb",
				kind:       game.WeaponCluckster,
				ammoType:   game.AmmoTypeEgg,
				wheelColor: color.FromBytes(25, 40, 120, 255),
			},
		},
	}
	for i := range weapons.weapons {
		weapons.weapons[i].init()
	}
}

func (weapons *Weapons) Layout(queue *ui.RenderQueue, deltaTime float32, stats PlayerStats) {
	// Update weapon wheel
	wheel := &weapons.weaponWheel
	if settings.Current.ActionWeaponWheel.JustPressed() {
		*wheel = newWeaponWheel(weapons.weapons[:])
	} else if settings.Current.ActionWeaponWheel.JustReleased() && stats.Health > 0 && weapons.Get(wheel.highlightedWeapon) != nil {
		weapons.Select(wheel.highlightedWeapon)
	}

	if stats.WeaponWheelOpenness > 0.0 {
		wheel.Layout(queue, stats.WeaponWheelOpenness)
	}

	// Transition selected weapon
	if weapon := weapons.Selected(); weapon == nil || !weapon.isSelected() {
		weapons.selectedWeapon = weapons.nextWeapon
		if weapons.selectedWeapon >= game.WeaponSickle {
			weapon = weapons.Selected()
			weapon.onSelect()
		}
	}

	// Update weapons
	for i := range weapons.weapons {
		wep := &weapons.weapons[i]
		wep.update(deltaTime, stats.MoveSpeed, stats.Ammo)
		if game.WeaponType(i) == weapons.selectedWeapon {
			queue.Add(&wep.sprite)
		}
	}
}

func (weapons *Weapons) Get(index game.WeaponType) *Weapon {
	if index == game.WeaponNone || index >= game.WeaponCount {
		return nil
	}
	return &weapons.weapons[index]
}

func (weapons *Weapons) Selected() *Weapon {
	return weapons.Get(weapons.selectedWeapon)
}

func (weapons *Weapons) AttemptFire(ammo *game.Ammo) bool {
	weapon := weapons.Selected()
	if weapon != nil && weapon.canFire(*ammo) {
		weapon.fire(ammo)
		return true
	}
	return false
}

func (weapons *Weapons) Select(order game.WeaponType) {
	wantWeapon := weapons.Get(order)
	if order == weapons.selectedWeapon {
		return
	}
	if order >= game.WeaponSickle && !wantWeapon.Equipped {
		weapons.hud.ShowMessage(settings.Localize(wantWeapon.name+"NotFound"), 2, color.Red)
		return
	}
	if weapons.selectedWeapon >= game.WeaponSickle {
		weapons.Selected().onDeselect()
	}
	weapons.nextWeapon = order
}

func (weapons *Weapons) ListEquipped() (result [game.WeaponCount]bool) {
	for i := range game.WeaponCount {
		if weap := weapons.Get(i); weap != nil && weap.Equipped {
			result[i] = true
		}
	}
	return
}
