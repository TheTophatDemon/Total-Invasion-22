package hud

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Weapons struct {
	weapons                    [WeaponCount]Weapon
	weaponWheel                WeaponWheel
	selectedWeapon, nextWeapon WeaponKind
}

func (weapons *Weapons) init() {
	*weapons = Weapons{
		weapons: [WeaponCount]Weapon{
			// Sickle
			WeaponSickle: {
				name:         settings.Localize("sickle"),
				kind:         WeaponSickle,
				cooldown:     0.25,
				texturePath:  "assets/textures/ui/sickle_hud.png",
				swayExtents:  mgl32.Vec2{32.0, 16.0},
				swaySpeed:    mgl32.Vec2{0.75, 1.5},
				ammoType:     game.AMMO_TYPE_SICKLE,
				ammoCost:     1,
				spriteOffset: mgl32.Vec2{settings.UIWidth() / 4.0, 0.0},
				updateFunc: func(sickle *Weapon, deltaTime float32, ammo game.Ammo) {
					throwAnim, _ := sickle.sprite.Texture.GetAnimation(fireAnimName)
					animPlayer := &sickle.sprite.AnimPlayer
					if sickle.canFire(ammo) && animPlayer.CurrentAnimation().Name == throwAnim.Name {
						catchAnim, _ := sickle.sprite.Texture.GetAnimation("catch")
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
			WeaponChicken: {
				name:          settings.Localize("chickenCannon"),
				kind:          WeaponChicken,
				cooldown:      0.15,
				texturePath:   "assets/textures/ui/chicken_cannon_hud.png",
				swayExtents:   mgl32.Vec2{16.0, 8.0},
				swaySpeed:     mgl32.Vec2{0.5, 1.0},
				ammoType:      game.AMMO_TYPE_EGG,
				ammoCost:      1,
				wheelColor:    color.FromBytes(0, 0, 255, 255),
				wheelIconPath: "assets/textures/sprites/chicken_cannon.png",
				noiseLevel:    20.0,
				isShooter:     true,
			},
			// Grenade launcher
			WeaponGrenade: {
				name:          settings.Localize("grenadeLauncher"),
				kind:          WeaponGrenade,
				cooldown:      1.0,
				texturePath:   "assets/textures/ui/grenade_launcher_hud.png",
				swayExtents:   mgl32.Vec2{16.0, 8.0},
				swaySpeed:     mgl32.Vec2{0.75, 1.25},
				ammoType:      game.AMMO_TYPE_GRENADE,
				ammoCost:      1,
				wheelColor:    color.FromBytes(0, 170, 0, 255),
				wheelIconPath: "assets/textures/sprites/grenade_launcher.png",
				noiseLevel:    50.0,
				isShooter:     true,
			},
			// Parusu
			WeaponParusu: {
				name:          settings.Localize("parusu"),
				kind:          WeaponParusu,
				cooldown:      0.075,
				texturePath:   "assets/textures/ui/parusu_hud.png",
				swayExtents:   mgl32.Vec2{16.0, 8.0},
				swaySpeed:     mgl32.Vec2{0.5, 1.0},
				ammoType:      game.AMMO_TYPE_PLASMA,
				ammoCost:      1,
				wheelColor:    color.FromBytes(0, 255, 130, 255),
				wheelIconPath: "assets/textures/sprites/parusu.png",
				noiseLevel:    40.0,
				isShooter:     true,
				updateFunc: func(parusu *Weapon, deltaTime float32, ammo game.Ammo) {
					animPlayer := &parusu.sprite.AnimPlayer
					if animPlayer.CurrentAnimation().Name == fireAnimName && animPlayer.IsAtEnd() {
						idleAnim, _ := parusu.sprite.Texture.GetAnimation(idleAnimName)
						animPlayer.PlayNewAnim(idleAnim)
					}
				},
			},
			// Double grenade launcher (NOT IMPLEMENTED)
			WeaponDblGrenade: {
				name:          settings.Localize("doubleGrenadeLauncher"),
				kind:          WeaponDblGrenade,
				wheelColor:    color.FromBytes(255, 130, 0, 255),
				wheelIconPath: "assets/textures/sprites/double_grenade_launcher.png",
			},
			// Sign of Madness (NOT IMPLEMENTED)
			WeaponSign: {
				name:          settings.Localize("signOfMadness"),
				kind:          WeaponSign,
				wheelColor:    color.FromBytes(170, 0, 0, 255),
				wheelIconPath: "assets/textures/sprites/sign_of_madness.png",
			},
			// Airhorn
			WeaponAirhorn: {
				name:          settings.Localize("airhorn"),
				kind:          WeaponAirhorn,
				cooldown:      0.0,
				texturePath:   "assets/textures/ui/airhorn_hud.png",
				swayExtents:   mgl32.Vec2{32.0, 16.0},
				swaySpeed:     mgl32.Vec2{0.75, 1.5},
				ammoType:      game.AMMO_TYPE_NONE,
				spriteOffset:  mgl32.Vec2{settings.UIWidth() / 6.0, 0.0},
				noiseLevel:    100.0,
				wheelColor:    color.FromBytes(255, 0, 0, 255),
				wheelIconPath: "assets/textures/sprites/airhorn.png",
				fireSoundPath: "assets/sounds/weapon/airhorn.wav",
				updateFunc: func(airhorn *Weapon, deltaTime float32, ammo game.Ammo) {
					if !airhorn.heldDown {
						idleAnim, _ := airhorn.sprite.Texture.GetAnimation(idleAnimName)
						airhorn.sprite.AnimPlayer.PlayNewAnim(idleAnim)
						if airhorn.voice.IsPlaying() && airhorn.voice.GetTime() < 800 {
							airhorn.voice.Seek(800)
						}
					}
				},
			},
			// Defenestrator (NOT IMPLEMENTED)
			WeaponDefenestrator: {
				name: settings.Localize("defenestrator"),
				kind: WeaponDefenestrator,
			},
			// Cluckster Bomb (NOT IMPLEMENTED)
			WeaponCluckster: {
				name: settings.Localize("clucksterBomb"),
				kind: WeaponCluckster,
			},
		},
	}
	for i := range weapons.weapons {
		weapons.weapons[i].init()
	}
	weapons.weaponWheel = newWeaponWheel(weapons.weapons[:])
}

func (weapons *Weapons) Layout(queue *ui.RenderQueue, deltaTime float32, stats PlayerStats) {
	// Update weapon wheel
	wheel := &weapons.weaponWheel
	if _, justPressed, justReleased := input.ActionPressStates(settings.ACTION_WEAPON_WHEEL); justPressed {
		*wheel = newWeaponWheel(weapons.weapons[:])
	} else if justReleased && stats.Health > 0 {
		weap := weapons.Get(wheel.highlightedWeapon)
		if weap != nil && weap.equipped {
			weapons.Select(wheel.highlightedWeapon)
		}
		input.TrapMouse()
	}
	if stats.WeaponWheelOpenness > 0.0 {
		wheel.Layout(queue, stats.WeaponWheelOpenness)
	}

	// Transition selected weapon
	if weapon := weapons.Selected(); weapon == nil || !weapon.isSelected() {
		weapons.selectedWeapon = weapons.nextWeapon
		if weapons.selectedWeapon >= WeaponSickle {
			weapon = weapons.Selected()
			weapon.onSelect()
		}
	}

	// Update weapons
	for i := range weapons.weapons {
		wep := &weapons.weapons[i]
		wep.update(deltaTime, stats.MoveSpeed, *stats.Ammo)
		if WeaponKind(i) == weapons.selectedWeapon {
			queue.Add(&wep.sprite)
		}
	}
}

func (weapons *Weapons) Get(index WeaponKind) *Weapon {
	if index == WeaponNone || index >= WeaponCount {
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

func (weapons *Weapons) Select(order WeaponKind) {
	wantWeapon := weapons.Get(order)
	if wantWeapon == nil || order == weapons.selectedWeapon || (order >= WeaponSickle && !wantWeapon.equipped) {
		return
	}
	if weapons.selectedWeapon >= WeaponSickle {
		weapons.Selected().onDeselect()
	}
	weapons.nextWeapon = order
}

func (weapons *Weapons) Equip(order WeaponKind) {
	weapon := weapons.Get(order)
	if weapon == nil {
		return
	}
	weapon.equipped = true
}
