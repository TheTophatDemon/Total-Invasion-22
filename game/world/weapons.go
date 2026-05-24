package world

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
)

type (
	WeaponState uint8
	// Holds static data for the weapon type
	WeaponDef struct {
		AmmoCost                   int // Amount subtracted from ammo after firing
		AmmoType                   game.AmmoType
		Cooldown                   float32
		FireSoundPath              string
		IdleAnimName, FireAnimName string
		Index                      game.WeaponIndex
		IsShooter                  bool
		Name                       string // Weapon name, same as localization key.
		NoiseLevel                 float32
		RestOffset                 mgl32.Vec2  // Offset from the bottom center of the screen where the sprite will be after equipping
		SwayExtents                mgl32.Vec2  // Defines a rectangle on screen within which the weapon will sway
		SwaySpeed                  mgl32.Vec2  // Defines the speed at which the weapon will sway in each axis
		WheelColor                 color.Color // Color of weapon wheel frame
		WheelIconPath              string      // Path to icon texture displayed in weapon wheel slot

		InitFunc   func(w *Weapon)
		UpdateFunc func(w *Weapon, deltaTime float32, ammo game.Ammo)
		FireFunc   func(w *Weapon, player *Player, deltaTime float32, justPressed bool)
	}
	// Represents dynamic state of the weapon
	Weapon struct {
		*WeaponDef
		Sprite        ui.Element
		SpriteTarget  mgl32.Vec2 // The position the sprite moves towards when it's equipped. Set after initialization.
		Equipped      bool
		HeldDown      bool        // True while the fire button is being held
		State         WeaponState // Describes the state of transitional animations.
		Voice         tdaudio.VoiceId
		CooldownTimer float32
		Sway          float32 // Value tracking the timeline of the sway animation.
	}
)

const (
	WeaponStateInactive WeaponState = iota
	WeaponStateIntro
	WeaponStateReady
	WeaponStateOutro
	WeaponStateCount
)

var (
	WeaponSickle = WeaponDef{
		Name:          "sickle",
		Cooldown:      0.25,
		SwayExtents:   mgl32.Vec2{32.0, 16.0},
		SwaySpeed:     mgl32.Vec2{0.75, 1.5},
		AmmoType:      game.AmmoTypeSickle,
		AmmoCost:      1,
		RestOffset:    mgl32.Vec2{320.0, 0.0},
		WheelColor:    color.FromBytes(138, 138, 138, 255),
		WheelIconPath: "assets/textures/sprites/sickle.png",
		NoiseLevel:    10.0,
		IsShooter:     true,
		IdleAnimName:  "idle",
		Index:         game.WeaponIndexSickle,
		FireAnimName:  "fire",
		InitFunc: func(sickle *Weapon) {
			sickleTex := cache.GetTexture("assets/textures/ui/sickle_hud.png")
			sickle.Sprite = ui.NewBox(ui.Transform{
				Position: mgl32.Vec2{320.0, 0.0},
				Origin:   ui.Ratios{0.5, 1.0},
				Anchor:   ui.Ratios{0.5, 1.0},
			}, sickleTex)
			fireAnim, _ := sickleTex.GetAnimation(sickle.FireAnimName)
			sickle.Sprite.AnimPlayer.PlayNewAnim(fireAnim)
		},
		UpdateFunc: func(sickle *Weapon, deltaTime float32, ammo game.Ammo) {
			animPlayer := &sickle.Sprite.AnimPlayer
			if sickle.CanFire(ammo) && animPlayer.CurrentAnimation().Name == sickle.FireAnimName {
				catchAnim, _ := sickle.Sprite.BgTexture.GetAnimation("catch")
				animPlayer.PlayNewAnim(catchAnim)
				cache.GetSfx("assets/sounds/weapon/sickle_return.wav").Play()
			}
		},
		FireFunc: func(sickle *Weapon, player *Player, deltaTime float32, justPressed bool) {
			firePos := player.Body().Position.Add(player.actor.FacingVec().Mul(0.5))
			SpawnSickle(firePos, player.actor.FacingVec(), player.id.Handle)
		},
	}
	WeaponChicken = WeaponDef{
		Name:          "chickenCannon",
		Cooldown:      0.15,
		SwayExtents:   mgl32.Vec2{16.0, 8.0},
		SwaySpeed:     mgl32.Vec2{0.5, 1.0},
		AmmoType:      game.AmmoTypeEgg,
		AmmoCost:      1,
		WheelColor:    color.FromBytes(0, 0, 255, 255),
		WheelIconPath: "assets/textures/sprites/chicken_cannon.png",
		NoiseLevel:    20.0,
		IsShooter:     true,
		IdleAnimName:  "idle",
		Index:         game.WeaponIndexChicken,
		FireAnimName:  "fire",
		FireSoundPath: "assets/sounds/weapon/chickengun.wav",
		InitFunc: func(cannon *Weapon) {
			cannonTex := cache.GetTexture("assets/textures/ui/chicken_cannon_hud.png")
			cannon.Sprite = ui.NewBox(ui.Transform{
				Origin: ui.Ratios{0.5, 1.0},
				Anchor: ui.Ratios{0.5, 1.0},
			}, cannonTex)
		},
		FireFunc: func(chicken *Weapon, player *Player, deltaTime float32, justPressed bool) {
			firePos := player.Body().Position.Add(player.actor.FacingVec().Mul(0.5).Add(mgl32.Vec3{0.0, -0.15, 0.0}))
			SpawnEgg(firePos, player.actor.FacingVec(), player.id.Handle)
		},
	}
	WeaponGrenade = WeaponDef{
		Name:          "grenadeLauncher",
		Cooldown:      1.0,
		SwayExtents:   mgl32.Vec2{16.0, 8.0},
		SwaySpeed:     mgl32.Vec2{0.75, 1.25},
		AmmoType:      game.AmmoTypeGrenade,
		AmmoCost:      1,
		WheelColor:    color.FromBytes(0, 170, 0, 255),
		WheelIconPath: "assets/textures/sprites/grenade_launcher.png",
		NoiseLevel:    50.0,
		IsShooter:     true,
		IdleAnimName:  "idle",
		Index:         game.WeaponIndexGrenade,
		FireAnimName:  "fire",
		FireSoundPath: "assets/sounds/weapon/grenadelaunch.wav",
		InitFunc: func(grenade *Weapon) {
			grenadeTex := cache.GetTexture("assets/textures/ui/grenade_launcher_hud.png")
			grenade.Sprite = ui.NewBox(ui.Transform{
				Origin: ui.Ratios{0.5, 1.0},
				Anchor: ui.Ratios{0.5, 1.0},
			}, grenadeTex)
		},
		FireFunc: func(grenade *Weapon, player *Player, deltaTime float32, justPressed bool) {
			firePos := player.Body().Position.Add(player.actor.FacingVec().Mul(1.0).Add(mgl32.Vec3{0.0, 0.15, 0.0}))
			SpawnGrenade(firePos, player.actor.FacingVec(), player.id.Handle)
		},
	}
	WeaponParusu = WeaponDef{
		Name:          "parusu",
		Cooldown:      0.075,
		SwayExtents:   mgl32.Vec2{16.0, 8.0},
		SwaySpeed:     mgl32.Vec2{0.5, 1.0},
		AmmoType:      game.AmmoTypePlasma,
		AmmoCost:      1,
		WheelColor:    color.FromBytes(0, 255, 130, 255),
		WheelIconPath: "assets/textures/sprites/parusu.png",
		NoiseLevel:    40.0,
		IsShooter:     true,
		IdleAnimName:  "idle",
		Index:         game.WeaponIndexParusu,
		FireAnimName:  "fire",
		FireSoundPath: "assets/sounds/weapon/parusu.wav",
		InitFunc: func(parusu *Weapon) {
			parusuTex := cache.GetTexture("assets/textures/ui/parusu_hud.png")
			parusu.Sprite = ui.NewBox(ui.Transform{
				Origin: ui.Ratios{0.5, 1.0},
				Anchor: ui.Ratios{0.5, 1.0},
			}, parusuTex)
		},
		UpdateFunc: func(parusu *Weapon, deltaTime float32, ammo game.Ammo) {
			animPlayer := &parusu.Sprite.AnimPlayer
			if animPlayer.CurrentAnimation().Name == parusu.FireAnimName && animPlayer.IsAtEnd() {
				idleAnim, _ := parusu.Sprite.BgTexture.GetAnimation(parusu.IdleAnimName)
				animPlayer.PlayNewAnim(idleAnim)
			}
		},
		FireFunc: func(parusu *Weapon, player *Player, deltaTime float32, justPressed bool) {
			firePos := player.Body().Position.Add(player.actor.FacingVec().Mul(0.5).Add(mgl32.Vec3{0.0, -0.25, 0.0}))
			SpawnPlasmaBall(firePos, player.actor.FacingVec(), player.id.Handle, false)
		},
	}
	WeaponDblGrenade = WeaponDef{
		// NOT IMPLEMENTED
		Name:          "doubleGrenadeLauncher",
		Index:         game.WeaponIndexDblGrenade,
		WheelColor:    color.FromBytes(255, 130, 0, 255),
		WheelIconPath: "assets/textures/sprites/double_grenade_launcher.png",
	}
	WeaponSign = WeaponDef{
		// NOT IMPLEMENTED
		Name:          "signOfMadness",
		Index:         game.WeaponIndexSign,
		WheelColor:    color.FromBytes(170, 0, 0, 255),
		WheelIconPath: "assets/textures/sprites/sign_of_madness.png",
	}
	WeaponAirhorn = WeaponDef{
		Name:          "airhorn",
		Cooldown:      0,
		SwayExtents:   mgl32.Vec2{32.0, 16.0},
		SwaySpeed:     mgl32.Vec2{0.75, 1.5},
		AmmoType:      game.AmmoTypeNone,
		NoiseLevel:    100.0,
		WheelColor:    color.FromBytes(255, 0, 0, 255),
		WheelIconPath: "assets/textures/sprites/airhorn.png",
		FireSoundPath: "assets/sounds/weapon/airhorn.wav",
		IdleAnimName:  "idle",
		Index:         game.WeaponIndexAirhorn,
		FireAnimName:  "fire",
		InitFunc: func(airhorn *Weapon) {
			airhornTex := cache.GetTexture("assets/textures/ui/airhorn_hud.png")
			airhorn.Sprite = ui.NewBox(ui.Transform{
				Position: mgl32.Vec2{213.0, 0.0},
				Origin:   ui.Ratios{0.5, 1.0},
				Anchor:   ui.Ratios{0.5, 1.0},
			}, airhornTex)
		},
		UpdateFunc: func(airhorn *Weapon, deltaTime float32, ammo game.Ammo) {
			if !airhorn.HeldDown {
				idleAnim, _ := airhorn.Sprite.BgTexture.GetAnimation(airhorn.IdleAnimName)
				airhorn.Sprite.AnimPlayer.PlayNewAnim(idleAnim)
				if airhorn.Voice.IsPlaying() && airhorn.Voice.GetTime() < 800 {
					airhorn.Voice.Seek(800)
				}
			}
		},
		FireFunc: func(airhorn *Weapon, player *Player, deltaTime float32, justPressed bool) {
			if justPressed {
				//TODO: It would probably be more efficient to use the BSP tree here
				enemyIter := gWorld.Enemies.Iter()
				for {
					enemy, _ := enemyIter.Next()
					if enemy == nil {
						break
					}
					if enemy.actor.Health > 0 && enemy.state != &enemy.stunState {
						diff := enemy.actor.Position().Sub(player.actor.Position())
						dist := diff.Len()
						if dist > 0.0 && dist < 3.0 && diff.Mul(1.0/dist).Dot(player.actor.FacingVec()) > 0.9 {
							enemy.OnDamage(player, 1.0)
							enemy.changeState(&enemy.stunState)
						}
					}
				}
			}
		},
	}
	WeaponDefenestrator = WeaponDef{
		Name:       "defenestrator",
		AmmoType:   game.AmmoTypePlasma,
		Index:      game.WeaponIndexDefenestrator,
		WheelColor: color.FromBytes(32, 32, 32, 255),
	}
	WeaponCluckster = WeaponDef{
		Name:       "clucksterBomb",
		AmmoType:   game.AmmoTypeEgg,
		Index:      game.WeaponIndexCluckster,
		WheelColor: color.FromBytes(25, 40, 120, 255),
	}
)

func (weap *Weapon) Init(weaponDef *WeaponDef, equipped bool) {
	*weap = Weapon{
		WeaponDef: weaponDef,
		Equipped:  equipped,
	}
	if weap.InitFunc != nil {
		weap.InitFunc(weap)
	}
	if weap.Sprite.BgTexture != nil {
		idleAnim, _ := weap.Sprite.BgTexture.GetAnimation(weap.IdleAnimName)
		if !weap.Sprite.AnimPlayer.IsPlaying() {
			weap.Sprite.AnimPlayer.PlayNewAnim(idleAnim)
		}
		weap.Sprite.SetSize(idleAnim.Frames[0].Rect.SizeVec().Mul(SpriteScale()))
	}

	weap.SpriteTarget = weap.Sprite.Position()
}

func (weap *Weapon) AttemptFire(player *Player, deltaTime float32, justPressed bool) bool {
	if weap == nil || weap.WeaponDef == nil || !weap.CanFire(player.ammo) {
		return false
	}
	weap.CooldownTimer = weap.Cooldown
	player.ammo[weap.AmmoType] -= weap.AmmoCost
	weap.HeldDown = true

	fireAnim, ok := weap.Sprite.BgTexture.GetAnimation(weap.FireAnimName)
	if !ok {
		log.Printf("fire anim not found for weapon %v\n", weap.Name)
		return false
	}

	if weap.Sprite.AnimPlayer.CurrentAnimation().Name != fireAnim.Name {
		weap.Sprite.AnimPlayer.PlayNewAnim(fireAnim)
		if len(weap.FireSoundPath) > 0 {
			weap.Voice = cache.GetSfx(weap.FireSoundPath).Play()
		}
	}

	if weap.FireFunc != nil {
		weap.FireFunc(weap, player, deltaTime, justPressed)
	}
	return true
}

func (weap *Weapon) OnSelect() {
	if weap == nil || weap.WeaponDef == nil {
		return
	}
	weap.Sway = 0.0
	weap.State = WeaponStateIntro
	idleAnim, _ := weap.Sprite.BgTexture.GetAnimation(weap.IdleAnimName)
	weap.Sprite.AnimPlayer.PlayIfNotAlready(idleAnim)
}

func (weap *Weapon) OnDeselect() {
	if weap.State != WeaponStateInactive {
		weap.State = WeaponStateOutro
	}
}

func (weap *Weapon) CanFire(ammo game.Ammo) bool {
	return weap.State == WeaponStateReady && weap.CooldownTimer <= 0.0 && ammo[weap.AmmoType] >= weap.AmmoCost
}

func (weap *Weapon) Update(deltaTime float32, swayAmount float32, ammo game.Ammo) {
	if weap == nil || weap.WeaponDef == nil {
		return
	}
	weap.Sprite.AnimPlayer.Update(deltaTime)
	weap.CooldownTimer = max(weap.CooldownTimer-deltaTime, 0.0)
	swayOfs := mgl32.Vec2{
		math2.Cos(weap.Sway*weap.SwaySpeed[0]) * weap.SwayExtents[0],
		math2.Sin(weap.Sway*weap.SwaySpeed[1]) * weap.SwayExtents[1],
	}
	endPos := weap.SpriteTarget.Add(swayOfs)
	startPos := endPos.Add(mgl32.Vec2{0.0, weap.Sprite.AnimPlayer.Frame().Rect.Height * SpriteScale()})
	switch weap.State {
	case WeaponStateReady:
		weap.Sway += deltaTime * swayAmount
		// Sway the weapon according to player movement
		weap.Sprite.SetPosition(endPos)
	case WeaponStateIntro, WeaponStateOutro:
		// Move the weapon towards its screen position.
		target := endPos
		if weap.State == WeaponStateOutro {
			target = startPos
		}
		diff := target.Sub(weap.Sprite.Position())
		dist := diff.Len()
		moveAmt := deltaTime * 3072.0
		if dist < moveAmt {
			weap.Sprite.SetPosition(target)
			weap.State = (weap.State + 1) % WeaponStateCount
		} else {
			weap.Sprite.Translate(diff.Mul(moveAmt / dist))
		}
	case WeaponStateInactive:
		weap.Sprite.SetPosition(startPos)
	}

	if weap.UpdateFunc != nil {
		weap.UpdateFunc(weap, deltaTime, ammo)
	} else {
		// Default behavior: Switch to idle animation when firing is complete.
		idleAnim, _ := weap.Sprite.BgTexture.GetAnimation(weap.IdleAnimName)
		if weap.CanFire(ammo) || ammo[weap.AmmoType] == 0 {
			weap.Sprite.AnimPlayer.ChangeAnimation(idleAnim)
		}
	}

	weap.HeldDown = false
}
