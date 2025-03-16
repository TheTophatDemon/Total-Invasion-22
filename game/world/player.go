package world

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game"

	"tophatdemon.com/total-invasion-ii/game/hud"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const (
	USE_DIST float32 = 3.0
)

const (
	OVERHEAL_RESTORE_RATE = 1.0
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
	keys                   game.KeyType
}

var _ HasActor = (*Player)(nil)
var _ comps.HasBody = (*Player)(nil)

func (player *Player) Actor() *Actor {
	return &player.actor
}

func (player *Player) Body() *comps.Body {
	return &player.actor.body
}

func SpawnPlayer(world *World, position, angles mgl32.Vec3, camera scene.Id[*Camera]) (id scene.Id[*Player], player *Player, err error) {
	id, player, err = world.Players.New()
	if err != nil {
		return
	}
	player.id = id
	player.initialCollisionLayers = COL_LAYER_ACTORS | COL_LAYER_PLAYERS
	player.actor = Actor{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAngles(
				position, angles,
			),
			Shape:  collision.NewSphere(0.7),
			Layer:  player.initialCollisionLayers,
			Filter: COL_FILTER_FOR_ACTORS,
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

	tex := cache.GetTexture("assets/textures/sprites/segan.png")
	player.Sprite = comps.NewSpriteRender(tex)
	winAnim, _ := tex.GetAnimation("victory")
	player.AnimPlayer = comps.NewAnimationPlayer(winAnim, false)

	// Initialize weapons
	player.ammo[game.AMMO_TYPE_NONE] = 0
	player.ammo[game.AMMO_TYPE_SICKLE] = 1
	player.world.Hud.EquipWeapon(hud.WEAPON_ORDER_SICKLE)
	player.world.Hud.SelectWeapon(hud.WEAPON_ORDER_SICKLE)

	return
}

func (player *Player) Update(deltaTime float32) {
	if player.world.InWinState() {
		// Win logic
		if !player.AnimPlayer.IsPlaying() {
			player.AnimPlayer.Play()
		}
		player.AnimPlayer.Update(deltaTime)
		player.world.Hud.SelectWeapon(hud.WEAPON_ORDER_NONE)
		player.actor.inputForward = 0.0
		player.actor.inputStrafe = 0.0
		player.transitionTimer += deltaTime
		if (player.transitionTimer > 2.0 && input.IsActionPressed(settings.ACTION_FIRE)) || player.transitionTimer > 35.0 {
			player.world.ChangeMap(player.world.nextLevel)
		}
	} else if player.actor.Health > 0 {
		if player.world.IsOnPlayerCamera() {
			player.takeUserInput(deltaTime)
		} else {
			player.actor.inputForward = 0
			player.actor.inputStrafe = 0
		}
		if player.actor.Health > player.actor.TargetHealth {
			// When overhealed, gradually decrease health back to base level
			player.actor.Health = math2.Clamp(player.actor.Health-OVERHEAL_RESTORE_RATE*deltaTime, player.actor.TargetHealth, player.actor.MaxHealth)
		}
		if camera, ok := player.Camera.Get(); ok {
			// Keep camera transform in sync with the player
			camera.Transform = player.Body().Transform
		}
	} else {
		// Death logic
		player.world.Hud.SelectWeapon(hud.WEAPON_ORDER_NONE)
		player.world.Hud.FlashScreen(color.Red.WithAlpha(0.5), 1.0)
		player.actor.inputForward = 0.0
		player.actor.inputStrafe = 0.0
		if camera, ok := player.Camera.Get(); ok {
			if camera.Transform.Rotation().X() > -math.Pi/4.0 {
				camera.Transform.Rotate(-deltaTime, 0.0, 0.0)
			}
			if camera.Transform.Position().Y()-player.actor.Position().Y() > -player.Body().Shape.(collision.Sphere).Radius() {
				player.cameraFall -= deltaTime * 10.0
				camera.Transform.Translate(0.0, deltaTime*player.cameraFall, 0.0)
			}
		}
		player.transitionTimer += deltaTime
		if (player.transitionTimer > 2.0 && input.IsActionPressed(settings.ACTION_FIRE)) || player.transitionTimer > 10.0 {
			player.world.ChangeMap(player.world.GameMap.Name())
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

	player.world.Hud.UpdatePlayerStats(deltaTime, hud.PlayerStats{
		Health:    int(player.actor.Health),
		Noclip:    player.Body().Layer == COL_LAYER_NONE,
		GodMode:   player.godMode,
		Ammo:      &player.ammo,
		Keys:      player.keys,
		MoveSpeed: player.actor.body.Velocity.Len(),
	})
}

func (player *Player) Render(context *render.Context) {
	if player.world.InWinState() {
		player.Sprite.Render(&player.Body().Transform, &player.AnimPlayer, context, player.actor.YawAngle)
	}
}

func (player *Player) takeUserInput(deltaTime float32) {
	_ = deltaTime
	if input.IsActionPressed(settings.ACTION_FORWARD) {
		player.actor.inputForward = 1.0
	} else if input.IsActionPressed(settings.ACTION_BACK) {
		player.actor.inputForward = -1.0
	} else {
		player.actor.inputForward = 0.0
	}

	if input.IsActionPressed(settings.ACTION_RIGHT) {
		player.actor.inputStrafe = 1.0
	} else if input.IsActionPressed(settings.ACTION_LEFT) {
		player.actor.inputStrafe = -1.0
	} else {
		player.actor.inputStrafe = 0.0
	}

	if input.IsActionJustPressed(settings.ACTION_NOCLIP) {
		var message string = settings.Localize("noclipActivate")
		if player.Body().Layer != COL_LAYER_NONE {
			player.Body().Layer = COL_LAYER_NONE
			player.Body().Filter = COL_LAYER_NONE
		} else {
			player.Body().Layer = player.initialCollisionLayers
			player.Body().Filter = COL_FILTER_FOR_ACTORS
			message = settings.Localize("noclipDeactivate")
		}
		player.world.Hud.ShowMessage(message, 4.0, 100, color.Red)
	}

	if input.IsActionJustPressed(settings.ACTION_GODMODE) {
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
		player.world.Hud.ShowMessage(message, 4.0, 100, color.Red)
	}

	if input.IsActionJustPressed(settings.ACTION_MARYSUE) {
		player.world.Hud.ShowMessage("Mary Sue mode activated!", 4.0, 100, color.Red)
		for i := hud.WEAPON_ORDER_SICKLE; i < hud.WEAPON_ORDER_COUNT; i++ {
			player.world.Hud.EquipWeapon(i)
		}
		for i := range player.ammo {
			player.ammo[i] = game.AmmoLimits[i]
		}
		player.keys = game.KEY_TYPE_ALL
	}

	if input.IsActionJustPressed(settings.ACTION_DIE) {
		player.actor.Health = 0
	}

	if input.IsActionJustPressed(settings.ACTION_USE) {
		rayOrigin := player.Body().Transform.Position()
		rayDir := player.Body().Transform.Forward()
		hit, closestBody := player.world.Raycast(rayOrigin, rayDir, COL_FILTER_FOR_ACTORS, USE_DIST, player)
		if hit.Hit && !closestBody.IsNil() {
			if usable, isUsable := scene.Get[Usable](closestBody); isUsable {
				usable.OnUse(player)
			}
		}
	}

	if input.IsActionJustPressed(settings.ACTION_SICKLE) {
		player.world.Hud.SelectWeapon(hud.WEAPON_ORDER_SICKLE)
	} else if input.IsActionJustPressed(settings.ACTION_CHICKEN) {
		player.world.Hud.SelectWeapon(hud.WEAPON_ORDER_CHICKEN)
	} else if input.IsActionJustPressed(settings.ACTION_GRENADE) {
		player.world.Hud.SelectWeapon(hud.WEAPON_ORDER_GRENADE)
	} else if input.IsActionJustPressed(settings.ACTION_PARUSU) {
		player.world.Hud.SelectWeapon(hud.WEAPON_ORDER_PARUSU)
	} else if input.IsActionJustPressed(settings.ACTION_AIRHORN) {
		player.world.Hud.SelectWeapon(hud.WEAPON_ORDER_AIRHORN)
	}

	if input.IsActionPressed(settings.ACTION_SLOW) {
		player.actor.MaxSpeed = player.WalkSpeed
	} else {
		player.actor.MaxSpeed = player.RunSpeed
	}

	if weap := player.world.Hud.SelectedWeapon(); weap != nil && input.IsActionPressed(settings.ACTION_FIRE) {
		var cast collision.RaycastResult
		if weap.IsShooter() {
			// Don't fire if there is a wall too close in front
			cast, _ = player.world.Raycast(player.Body().Transform.Position(), player.Body().Transform.Forward(), COL_LAYER_MAP, 1.5, player)
		}

		if !cast.Hit && player.world.Hud.AttemptFireWeapon(&player.ammo) {
			player.AttackWithWeapon(input.IsActionJustPressed(settings.ACTION_FIRE))
		}
	}

	player.actor.YawAngle -= input.ActionAxis(settings.ACTION_LOOK_HORZ)
}

func (player *Player) ProcessSignal(signal any) {
	switch signal.(type) {
	case game.TeleportationSignal:
		player.world.Hud.FlashScreen(color.Color{R: 1.0, G: 0.0, B: 1.0, A: 1.0}, 2.0)
	}
}

func (player *Player) OnDamage(sourceEntity any, damage float32) bool {
	if player.godMode {
		return false
	}
	player.actor.Health = max(0, player.actor.Health-damage)

	if player.actor.Health > 0 {
		player.world.Hud.FlashScreen(color.Red.WithAlpha(0.5), 1.0)
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
				player.world.Hud.SuggestPlayerFace(hud.FaceStateHurtFront)
			} else if forward.Cross(dmgDir).Y() > 0.0 {
				// Source is to the left
				player.world.Hud.SuggestPlayerFace(hud.FaceStateHurtLeft)
			} else {
				// Source is to the right
				player.world.Hud.SuggestPlayerFace(hud.FaceStateHurtRight)
			}
		} else {
			player.world.Hud.SuggestPlayerFace(hud.FaceStateHurtFront)
		}
	}
	return true
}

// Adds ammo to the player's amounts, checking the limits to not overfill. Returns false if player has max ammo already.
func (player *Player) AddAmmo(ammoType game.AmmoType, amount int) bool {
	limit := game.AmmoLimits[ammoType]
	if player.ammo[ammoType] == limit {
		return false
	}
	newAmmo := player.ammo[ammoType] + amount
	if newAmmo > limit {
		player.ammo[ammoType] = limit
	} else {
		player.ammo[ammoType] = newAmmo
	}
	return true
}
