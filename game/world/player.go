package world

import (
	"math"
	"math/rand"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/timer"
	"tophatdemon.com/total-invasion-ii/game"

	"tophatdemon.com/total-invasion-ii/game/hud"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Player struct {
	Camera                                   scene.Id[*Camera]
	Sprite                                   comps.SpriteRender // Mainly shown during the victory state
	AnimPlayer                               comps.AnimationPlayer
	RunSpeed, WalkSpeed                      float32
	StandFriction, WalkFriction, RunFriction float32
	id                                       scene.Id[*Player]
	actor                                    Actor
	world                                    *World

	initialCollisionLayers collision.Mask
	cameraFall             float32 // Used to track the Y velocity of the camera as it falls to the ground after player death.
	transitionTimer        float32 // Counts the seconds until the game resets after winning or dying.
	godMode                bool    // If true, the player does not take damage.
	ammo                   game.Ammo
	keys                   game.Keys
	armorType              game.ArmorType
	armorAmount            float32
	weaponWheelOpenness    float32 // 1 if wheel is open, gradually drops to 0 after closing.
	punTimer               timer.Timer
	puns                   []string
}

var _ HasActor = (*Player)(nil)
var _ comps.HasBody = (*Player)(nil)

var actionToWeapon = map[input.Action]game.WeaponType{
	settings.ActionSickle:        game.WeaponSickle,
	settings.ActionChicken:       game.WeaponChicken,
	settings.ActionGrenade:       game.WeaponGrenade,
	settings.ActionParusu:        game.WeaponParusu,
	settings.ActionDblGrenade:    game.WeaponDblGrenade,
	settings.ActionSign:          game.WeaponSign,
	settings.ActionAirhorn:       game.WeaponAirhorn,
	settings.ActionDefenestrator: game.WeaponDefenestrator,
	settings.ActionCluckster:     game.WeaponCluckster,
}

func (player *Player) Actor() *Actor {
	return &player.actor
}

func (player *Player) Body() *comps.Body {
	return &player.actor.body
}

func SpawnPlayer(
	world *World,
	position,
	angles mgl32.Vec3,
	camera scene.Id[*Camera],
	changeInfo game.MapChangeSignal,
) (id scene.Id[*Player], player *Player, err error) {
	id, player, err = world.Players.New()
	if err != nil {
		return
	}
	player.id = id
	player.initialCollisionLayers = ColLayerActors | ColLayerPlayers
	player.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAngles(
				position, angles,
			),
			Shape:  collision.NewCylinder(0.7, 0.7),
			Layer:  player.initialCollisionLayers,
			Filter: ColFilterForActors,
			LockY:  true,
		},
		YawAngle:     mgl32.DegToRad(angles[1]),
		AccelRate:    100.0,
		Friction:     20.0,
		MaxHealth:    200,
		TargetHealth: 100,
		Health:       100,
		world:        world,
	}
	player.Camera = camera
	player.RunSpeed = 12.0
	player.WalkSpeed = 7.0
	player.StandFriction = 80.0
	player.WalkFriction = 1.0
	player.RunFriction = 20.0
	player.world = world
	player.cameraFall = 2.0

	player.punTimer = timer.Timer{
		Interval: 10.0,
	}

	player.puns = strings.Split(settings.Localize("puns"), "\n")
	rand.Shuffle(len(player.puns), func(i, j int) {
		player.puns[i], player.puns[j] = player.puns[j], player.puns[i]
	})

	tex := cache.GetTexture("assets/textures/sprites/segan.png")
	player.Sprite = comps.NewSpriteRender(tex)
	winAnim, _ := tex.GetAnimation("victory")
	player.AnimPlayer = comps.NewAnimationPlayer(winAnim, false)

	player.Body().Transform.SetRotation(0.0, player.actor.YawAngle, 0.0)

	// Initialize armor and ammo
	player.ammo = changeInfo.GiveAmmo
	player.ammo[game.AmmoTypeSickle] = 0
	player.armorType = changeInfo.GiveArmor
	player.armorAmount = changeInfo.ArmorAmount

	// Initialize weapons
	player.world.Hud.Weapons.Get(game.WeaponSickle).Equipped = true
	player.world.Hud.Weapons.Select(game.WeaponSickle)
	for i, equipped := range changeInfo.EquippedWeapons {
		if equipped {
			player.world.Hud.Weapons.Get(game.WeaponType(i)).Equipped = true
		}
	}

	if world.Hud.Intro.TimeLeft() > 0.0 {
		// Spawn intro sickle
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, 0.0, -80.0}, player.Body().Transform.Matrix())
		SpawnIntroSickle(player.world, firePos, player.Body().Transform.Rotation(), player.id.Handle)
	} else {
		player.ammo[game.AmmoTypeSickle] = 1
	}

	return
}

func (player *Player) Update(deltaTime float32) {
	hudPtr := &player.world.Hud

	player.weaponWheelOpenness = max(0.0, player.weaponWheelOpenness-(deltaTime*10.0))
	if hudPtr.Intro.TimeLeft() > 0.5 {
		// Wait
	} else if player.world.InWinState() {
		// Win logic
		if !player.AnimPlayer.IsPlaying() {
			player.AnimPlayer.Play()
		}
		player.AnimPlayer.Update(deltaTime)
		hudPtr.Weapons.Select(game.WeaponNone)
		player.actor.inputForward = 0.0
		player.actor.inputStrafe = 0.0
		player.transitionTimer += deltaTime
		if (player.transitionTimer > 2.0 && input.IsActionPressed(settings.ActionFire)) || player.transitionTimer > 35.0 {
			player.world.app.ProcessSignal(game.MapChangeSignal{
				NextMapPath:     player.world.impendingLevel,
				GiveAmmo:        player.ammo,
				GiveArmor:       player.armorType,
				ArmorAmount:     player.armorAmount,
				EquippedWeapons: hudPtr.Weapons.ListEquipped(),
			})
		}
	} else if player.actor.Health > 0 {
		if player.world.IsOnPlayerCamera() {
			player.takeUserInput(deltaTime)
		} else {
			player.punTimer.Reset()
			player.actor.inputForward = 0
			player.actor.inputStrafe = 0
		}

		if player.actor.Health > player.actor.TargetHealth {
			// When overhealed, gradually decrease health back to base level
			const overhealRate = 1.0
			player.actor.Health = math2.Clamp(player.actor.Health-overhealRate*deltaTime, player.actor.TargetHealth, player.actor.MaxHealth)
		}

		// Update puns
		if player.punTimer.Update(deltaTime) && len(player.puns) > 0 {
			pun := player.puns[0]
			hudPtr.ShowMessage(pun, 2.0, 10, color.Red)
			player.puns = player.puns[1:]
			player.punTimer.Elapsed -= 10.0
		}

		if camera, ok := player.Camera.Get(); ok {
			// Keep camera transform in sync with the player
			camera.Transform = player.Body().Transform
		}
	} else {
		// Death logic
		hudPtr.Weapons.Select(game.WeaponNone)
		player.armorType = game.ArmorTypeNone
		player.armorAmount = 0.0
		hudPtr.FlashScreen(color.Red.WithAlpha(0.5), 1.0)
		player.actor.inputForward = 0.0
		player.actor.inputStrafe = 0.0
		if camera, ok := player.Camera.Get(); ok {
			if camera.Transform.Rotation().X() > -math.Pi/4.0 {
				camera.Transform.Rotate(-deltaTime, 0.0, 0.0)
			}
			if camera.Transform.Position().Y()-player.actor.Position().Y() > -player.Body().Shape.Radius() {
				player.cameraFall -= deltaTime * 10.0
				camera.Transform.Translate(0.0, deltaTime*player.cameraFall, 0.0)
			}
		}
		player.transitionTimer += deltaTime
		if (player.transitionTimer > 2.0 && input.IsActionPressed(settings.ActionFire)) || player.transitionTimer > 10.0 {
			player.world.app.ProcessSignal(game.MapChangeSignal{NextMapPath: player.world.GameMap.Name})
		}
	}

	if math2.Abs(player.actor.inputForward) > mgl32.Epsilon || math2.Abs(player.actor.inputStrafe) > mgl32.Epsilon {
		if player.actor.MaxSpeed == player.WalkSpeed {
			player.actor.Friction = player.WalkFriction
		} else {
			player.actor.Friction = player.RunFriction
		}
	} else {
		player.actor.Friction = player.StandFriction
	}

	player.Body().Transform.SetRotation(0.0, player.actor.YawAngle, 0.0)
	player.actor.Update(deltaTime)

	hudPtr.PlayerStats = hud.PlayerStats{
		// Health needs to be rounded up so the face logic stays in sync with the player's state when the health reaches 0.
		Health:              int(math2.Ceil(player.actor.Health)),
		Noclip:              player.Body().Layer == ColLayerNone,
		GodMode:             player.godMode,
		Ammo:                player.ammo,
		Keys:                player.keys,
		MoveSpeed:           player.actor.body.Velocity.Len(),
		Armor:               player.armorType,
		ArmorAmount:         int(math2.Ceil(player.armorAmount)),
		WeaponWheelOpenness: player.weaponWheelOpenness,
	}
}

func (player *Player) Render(context *render.Context) {
	if player.world.InWinState() {
		player.Sprite.Render(&player.Body().Transform, &player.AnimPlayer, context, player.actor.YawAngle)
	}
}

func (player *Player) takeUserInput(deltaTime float32) {
	hudPtr := &player.world.Hud

	_ = deltaTime
	if input.IsActionPressed(settings.ActionForward) {
		player.actor.inputForward = 1.0
	} else if input.IsActionPressed(settings.ActionBack) {
		player.actor.inputForward = -1.0
	} else {
		player.actor.inputForward = 0.0
	}

	if input.IsActionPressed(settings.ActionRight) {
		player.actor.inputStrafe = 1.0
	} else if input.IsActionPressed(settings.ActionLeft) {
		player.actor.inputStrafe = -1.0
	} else {
		player.actor.inputStrafe = 0.0
	}

	if math2.Abs(player.actor.inputForward)+math2.Abs(player.actor.inputStrafe) > 0 {
		player.punTimer.Reset()
	}

	// Cheat codes
	if input.IsActionJustPressed(settings.ActionNoclip) {
		var message string = settings.Localize("noclipActivate")
		if player.Body().Layer != ColLayerNone {
			player.Body().Layer = ColLayerNone
			player.Body().Filter = ColLayerNone
		} else {
			player.Body().Layer = player.initialCollisionLayers
			player.Body().Filter = ColFilterForActors
			message = settings.Localize("noclipDeactivate")
		}
		hudPtr.ShowMessage(message, 4.0, 100, color.Red)
	}

	if input.IsActionJustPressed(settings.ActionGodMode) {
		if !player.godMode {
			player.actor.Health = player.actor.MaxHealth
		}
		player.godMode = !player.godMode
		var message string
		if player.godMode {
			message = settings.Localize("godModeActivate")
		} else {
			message = settings.Localize("godModeDeactivate")
		}
		hudPtr.ShowMessage(message, 4.0, 100, color.Red)
	}

	if input.IsActionJustPressed(settings.ActionMarySue) {
		hudPtr.ShowMessage("Mary Sue mode activated!", 4.0, 100, color.Red)
		for i := range game.WeaponCount - 1 {
			hudPtr.Weapons.Get(i + 1).Equipped = true
		}
		for i := range player.ammo {
			player.ammo[i] = game.AmmoType(i).Limit()
		}
		player.keys = game.KeysAll
	}

	if input.IsActionJustPressed(settings.ActionDie) {
		player.actor.Health = 0
	}

	if input.IsActionJustPressed(settings.ActionCastBlessing) {
		SpawnBlessing(player.world, player.actor.Position(), mgl32.Vec3{0.0, player.actor.YawAngle, 0.0}, player.id.Handle)
	}

	// Use key
	if input.IsActionJustPressed(settings.ActionUse) {
		rayOrigin := player.Body().Transform.Position()
		rayDir := player.Body().Transform.Forward()
		const useDist float32 = 3.0
		hit, closestBody := player.world.Raycast(rayOrigin, rayDir, ColFilterForActors, useDist, player)
		if hit.Hit && !closestBody.IsNil() {
			if usable, isUsable := scene.Get[Usable](closestBody); isUsable {
				usable.OnUse(player)
			}
		}
	}

	// Weapon selection
	for action, weapon := range actionToWeapon {
		if input.IsActionJustPressed(action) {
			hudPtr.Weapons.Select(weapon)
		}
	}

	if weap := hudPtr.Weapons.Selected(); weap != nil && input.IsActionPressed(settings.ActionFire) {
		var cast collision.Result
		if weap.IsShooter() {
			// Don't fire if there is a wall too close in front
			cast, _ = player.world.Raycast(player.Body().Transform.Position(), player.Body().Transform.Forward(), ColLayerMap, 1.5, player)
		}

		ammoBefore := player.ammo
		if !cast.Hit && hudPtr.Weapons.AttemptFire(&player.ammo) {
			player.AttackWithWeapon(input.IsActionJustPressed(settings.ActionFire))
			if player.armorType == game.ArmorTypeBullet && weap.Kind() != game.WeaponSickle {
				player.ammo = ammoBefore
			}
		}
	}

	// Sprinting
	if input.IsActionPressed(settings.ActionSlow) {
		player.actor.MaxSpeed = player.WalkSpeed
	} else {
		player.actor.MaxSpeed = player.RunSpeed
	}

	if input.IsActionPressed(settings.ActionWeaponWheel) {
		player.weaponWheelOpenness = 1.0
	}
	if player.weaponWheelOpenness <= 0.0 {
		player.actor.YawAngle -= input.ActionAxis(settings.ActionLookHorz)
	}
}

func (player *Player) ProcessSignal(signal any) {
	switch signal.(type) {
	case game.TeleportationSignal:
		player.world.Hud.FlashScreen(color.Color{R: 1.0, G: 0.0, B: 1.0, A: 1.0}, 2.0)
	}
}

func (player *Player) onIntersect(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	if item, isItem := otherEnt.(*Item); isItem {
		item.OnUse(player)
	}
}

func (player *Player) OnDamage(sourceEntity any, damage float32) bool {
	if player.godMode {
		return false
	}

	hudPtr := &player.world.Hud

	wasNonZero := player.armorAmount > 0
	player.armorAmount = max(0, player.armorAmount-damage)
	if player.armorAmount <= 0 && wasNonZero {
		player.armorType = game.ArmorTypeNone
		cache.GetSfx("assets/sounds/armor_break.wav").Play()
		hudPtr.ShowMessage(settings.Localize("armorBroken"), 2.0, 10, color.Red)
	}
	damage *= (1.0 - player.armorType.Defense())

	player.actor.Health = max(0, player.actor.Health-damage)

	if player.actor.Health > 0 {
		hudPtr.FlashScreen(color.Red.WithAlpha(0.5), 1.0)

		if bodyHaver, ok := sourceEntity.(comps.HasBody); ok {
			// Change the hurt face with respect to the direction the damage is coming from
			dmgDir := bodyHaver.Body().Transform.Position().Sub(player.Body().Transform.Position())
			if dmgDir.LenSqr() > 0.0 {
				dmgDir = dmgDir.Normalize()
			}
			forward := player.actor.FacingVec()
			halfFov := mgl32.DegToRad(settings.Current.Fov / 2.0)
			if angleTo := math2.Acos(dmgDir.Dot(forward)); angleTo < halfFov || angleTo > math.Pi-halfFov {
				// Source is in front or back
				hudPtr.StatusBar.SuggestPlayerFace(hud.FaceStateHurtFront)
			} else if forward.Cross(dmgDir).Y() > 0.0 {
				// Source is to the left
				hudPtr.StatusBar.SuggestPlayerFace(hud.FaceStateHurtLeft)
			} else {
				// Source is to the right
				hudPtr.StatusBar.SuggestPlayerFace(hud.FaceStateHurtRight)
			}
		} else {
			hudPtr.StatusBar.SuggestPlayerFace(hud.FaceStateHurtFront)
		}
	}
	return true
}

// Adds ammo to the player's amounts, checking the limits to not overfill. Returns false if player has max ammo already.
func (player *Player) AddAmmo(ammoType game.AmmoType, amount int) bool {
	limit := ammoType.Limit()
	if player.ammo[ammoType] == limit {
		return false
	}
	newAmmo := player.ammo[ammoType] + amount
	player.ammo[ammoType] = min(newAmmo, limit)
	return true
}

// Adds armor to the player's stats, checking to not overfill limits. Returns false if player has max armor already.
func (player *Player) AddArmor(armorType game.ArmorType, amount int) bool {
	if armorType == player.armorType && player.armorAmount >= game.MaxArmorAmount {
		return false
	}
	player.armorType = armorType
	player.armorAmount = min(player.armorAmount+float32(amount), game.MaxArmorAmount)
	return true
}

func (player *Player) AttackWithWeapon(justPressed bool) {
	player.punTimer.Reset()
	weapon := player.world.Hud.Weapons.Selected()
	if weapon == nil {
		return
	}
	switch weapon.Kind() {
	case game.WeaponSickle:
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, 0.0, -0.5}, player.Body().Transform.Matrix())
		SpawnSickle(player.world, firePos, player.Body().Transform.Rotation(), player.id.Handle)
	case game.WeaponChicken:
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, -0.15, -0.5}, player.Body().Transform.Matrix())
		SpawnEgg(player.world, firePos, player.Body().Transform.Rotation(), player.id.Handle)
		cache.GetSfx("assets/sounds/weapon/chickengun.wav").Play()
	case game.WeaponGrenade:
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, 0.15, -1.25}, player.Body().Transform.Matrix())
		SpawnGrenade(player.world, firePos, player.Body().Transform.Forward())
		cache.GetSfx("assets/sounds/weapon/grenadelaunch.wav").Play()
	case game.WeaponParusu:
		firePos := mgl32.TransformCoordinate(mgl32.Vec3{0.0, -0.25, -0.5}, player.Body().Transform.Matrix())
		SpawnPlasmaBall(player.world, firePos, player.Body().Transform.Rotation(), player.id.Handle, false)
		cache.GetSfx("assets/sounds/weapon/parusu.wav").Play()
	case game.WeaponAirhorn:
		if justPressed {
			enemyIter := player.world.Enemies.Iter()
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
	}
	player.actor.noisyTimer = 0.5
}
