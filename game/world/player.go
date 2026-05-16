package world

import (
	"math"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
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

	cameraFall                                                                            float32 // Used to track the Y velocity of the camera as it falls to the ground after player death.
	transitionTimer                                                                       float32 // Counts the seconds until the game resets after winning or dying.
	godMode                                                                               bool    // If true, the player does not take damage.
	ammo                                                                                  game.Ammo
	keys                                                                                  game.Keys
	armorType                                                                             game.ArmorType
	armorAmount                                                                           float32
	weaponWheelOpenness                                                                   float32 // 1 if wheel is open, gradually drops to 0 after closing.
	SelectedWeapon                                                                        *Weapon
	Sickle, Chicken, Grenade, Parusu, DblGrenade, Sign, Airhorn, Defenestrator, Cluckster Weapon
	punTimer                                                                              timer.Timer
	puns                                                                                  []string
	safety                                                                                bool // Prevents firing accidentally after exiting a menu
}

var _ HasActor = (*Player)(nil)
var _ comps.HasBody = (*Player)(nil)

func (player *Player) Actor() *Actor {
	return &player.actor
}

func (player *Player) Body() *comps.Body {
	return &player.actor.body
}

func SpawnPlayer(
	position,
	angles mgl32.Vec3,
	camera scene.Id[*Camera],
	changeInfo game.MapChangeSignal,
) (id scene.Id[*Player], player *Player, err error) {
	id, player, err = gWorld.Players.New()
	if err != nil {
		return
	}
	player.id = id
	player.actor = Actor{
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.6, 0.7, 0.6),
			Layers:   ColLayerActors | ColLayerPlayers,
		},
		collisionFilter: ColLayerMap | ColLayerActors | ColLayerInvisible,
		YawAngle:        mgl32.DegToRad(angles[1]),
		AccelRate:       100.0,
		Friction:        20.0,
		MaxHealth:       200,
		TargetHealth:    100,
		Health:          100,
	}
	player.Camera = camera
	player.RunSpeed = 12.0
	player.WalkSpeed = 7.0
	player.StandFriction = 80.0
	player.WalkFriction = 1.0
	player.RunFriction = 20.0
	player.cameraFall = 2.0

	player.punTimer = timer.Timer{
		Interval: 10.0,
	}

	player.puns = strings.Split(settings.Localize("puns"), "\n")
	rand.Shuffle(len(player.puns), func(i, j int) {
		player.puns[i], player.puns[j] = player.puns[j], player.puns[i]
	})

	tex := cache.GetTexture("assets/textures/sprites/segan.png")
	player.Sprite = comps.NewSpriteRender(tex, nil, nil)
	winAnim, _ := tex.GetAnimation("victory")
	player.AnimPlayer = comps.NewAnimationPlayer(winAnim, false)

	// Initialize armor and ammo
	player.ammo = changeInfo.GiveAmmo
	player.ammo[game.AmmoTypeSickle] = 0
	player.armorType = changeInfo.GiveArmor
	player.armorAmount = changeInfo.ArmorAmount

	// Initialize weapons
	player.Sickle.Init(&WeaponSickle, true)
	player.SelectedWeapon = &player.Sickle
	player.Chicken.Init(&WeaponChicken, changeInfo.EquippedChicken)
	player.Grenade.Init(&WeaponGrenade, changeInfo.EquippedGrenade)
	player.Parusu.Init(&WeaponParusu, changeInfo.EquippedParusu)
	player.DblGrenade.Init(&WeaponDblGrenade, changeInfo.EquippedDblGrenade)
	player.Sign.Init(&WeaponSign, changeInfo.EquippedSign)
	player.Airhorn.Init(&WeaponAirhorn, changeInfo.EquippedAirhorn)
	player.Defenestrator.Init(&WeaponDefenestrator, changeInfo.EquippedDefenestrator)
	player.Cluckster.Init(&WeaponCluckster, changeInfo.EquippedCluckster)

	if gWorld.Hud.Intro.TimeLeft() > 0.0 {
		// Spawn intro sickle
		firePos := player.Body().Position.Add(player.actor.FacingVec().Mul(80.0))
		SpawnIntroSickle(firePos, player.actor.FacingVec(), player.id.Handle)
	} else {
		player.ammo[game.AmmoTypeSickle] = 1
	}

	return
}

func (player *Player) Update(deltaTime float32) {
	hudPtr := &gWorld.Hud

	player.weaponWheelOpenness = max(0.0, player.weaponWheelOpenness-(deltaTime*10.0))
	if hudPtr.Intro.TimeLeft() > 0.5 {
		// Wait
	} else if gWorld.InWinState() {
		// Win logic
		if !player.AnimPlayer.IsPlaying() {
			player.AnimPlayer.Play()
		}
		player.AnimPlayer.Update(deltaTime)
		player.SelectedWeapon = nil
		player.actor.inputForward = 0.0
		player.actor.inputStrafe = 0.0
		player.transitionTimer += deltaTime
		if (player.transitionTimer > 2.0 && input.IsAnythingPressed()) || player.transitionTimer > 35.0 {
			gWorld.app.ProcessSignal(game.MapChangeSignal{
				NextMapPath:           gWorld.impendingLevel,
				GiveAmmo:              player.ammo,
				GiveArmor:             player.armorType,
				ArmorAmount:           player.armorAmount,
				EquippedChicken:       player.Chicken.Equipped,
				EquippedGrenade:       player.Grenade.Equipped,
				EquippedParusu:        player.Parusu.Equipped,
				EquippedDblGrenade:    player.DblGrenade.Equipped,
				EquippedSign:          player.Sign.Equipped,
				EquippedAirhorn:       player.Airhorn.Equipped,
				EquippedDefenestrator: player.Defenestrator.Equipped,
				EquippedCluckster:     player.Cluckster.Equipped,
			})
		}
	} else if player.actor.Health > 0 {
		if gWorld.IsOnPlayerCamera() {
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
			hudPtr.ShowMessage(pun, 10, color.Red)
			player.puns = player.puns[1:]
			player.punTimer.Elapsed -= 10.0
		}

		if camera, ok := player.Camera.Get(); ok {
			// Keep camera transform in sync with the player
			camera.Transform.SetPositionV(player.Body().Position)
			camera.Transform.SetRotation(0.0, player.actor.YawAngle, 0.0)
		}
	} else {
		// Death logic
		player.SelectedWeapon = nil
		player.armorType = game.ArmorTypeNone
		player.armorAmount = 0.0
		hudPtr.FlashScreen(color.Red.WithAlpha(0.5), 1.0)
		player.actor.inputForward = 0.0
		player.actor.inputStrafe = 0.0
		if camera, ok := player.Camera.Get(); ok {
			if camera.Transform.Rotation().X() > -math.Pi/4.0 {
				camera.Transform.Rotate(-deltaTime, 0.0, 0.0)
			}
			if camera.Transform.Position().Y()-player.actor.Position().Y() > -player.Body().Shape.Extents().LongestDimension()/2.0 {
				player.cameraFall -= deltaTime * 10.0
				camera.Transform.Translate(0.0, deltaTime*player.cameraFall, 0.0)
			}
		}
		player.transitionTimer += deltaTime
		if (player.transitionTimer > 2.0 && input.IsAnythingPressed()) || player.transitionTimer > 10.0 {
			gWorld.app.ProcessSignal(game.MapChangeSignal{NextMapPath: gWorld.GameMap.Name})
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

	player.actor.Update(deltaTime)

	//TODO: Move HUD into world package so we don't have to deal with this bullshit
	// hudPtr.PlayerStats = hud.PlayerStats{
	// 	// Health needs to be rounded up so the face logic stays in sync with the player's state when the health reaches 0.
	// 	Health:              int(math2.Ceil(player.actor.Health)),
	// 	Noclip:              player.actor.NoClip,
	// 	GodMode:             player.godMode,
	// 	Ammo:                player.ammo,
	// 	Keys:                player.keys,
	// 	MoveSpeed:           player.actor.body.Velocity.Len(),
	// 	Armor:               player.armorType,
	// 	ArmorAmount:         int(math2.Ceil(player.armorAmount)),
	// 	WeaponWheelOpenness: player.weaponWheelOpenness,
	// 	EquippedWeapons:     player.equippedWeapons,
	// 	SelectedWeapon:      player.SelectedWeapon,
	// }
}

func (player *Player) Render(context *render.Context) {
	if gWorld.InWinState() {
		player.Sprite.Render(player.Body().Position, &player.AnimPlayer, context, player.actor.YawAngle, string(settings.Current.Locale))
	}
}

func (player *Player) takeUserInput(deltaTime float32) {
	hudPtr := &gWorld.Hud

	_ = deltaTime
	if settings.Current.ActionForward.Pressed() {
		player.actor.inputForward = 1.0
	} else if settings.Current.ActionBack.Pressed() {
		player.actor.inputForward = -1.0
	} else {
		player.actor.inputForward = 0.0
	}

	if settings.Current.ActionRight.Pressed() {
		player.actor.inputStrafe = 1.0
	} else if settings.Current.ActionLeft.Pressed() {
		player.actor.inputStrafe = -1.0
	} else {
		player.actor.inputStrafe = 0.0
	}

	if math2.Abs(player.actor.inputForward)+math2.Abs(player.actor.inputStrafe) > 0 {
		player.punTimer.Reset()
	}

	// Cheat codes
	if settings.ActionNoclip.JustPressed() {
		if !player.actor.NoClip {
			player.actor.Body().ExcludeLayers(collision.MaskAll)
			player.actor.NoClip = true
			hudPtr.ShowMessage(settings.Localize("noclipActivate"), 100, color.Red)
		} else {
			player.actor.Body().RestoreLayers()
			player.actor.NoClip = false
			hudPtr.ShowMessage(settings.Localize("noclipDeactivate"), 100, color.Red)
		}
	}

	// God mode
	if settings.ActionGodMode.JustPressed() {
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
		hudPtr.ShowMessage(message, 100, color.Red)
	}

	// Mary sue mode
	if settings.ActionMarySue.JustPressed() {
		hudPtr.ShowMessage("Mary Sue mode activated!", 100, color.Red)
		player.Chicken.Equipped = true
		player.Grenade.Equipped = true
		player.Parusu.Equipped = true
		player.DblGrenade.Equipped = true
		player.Sign.Equipped = true
		player.Airhorn.Equipped = true
		player.Defenestrator.Equipped = true
		player.Cluckster.Equipped = true
		for i := range player.ammo {
			player.ammo[i] = game.AmmoType(i).Limit()
		}
		player.keys = game.KeysAll
	}

	// Unalive
	if settings.ActionDie.JustPressed() {
		player.actor.Health = 0
	}

	// Cast blessing
	if settings.ActionCastBlessing.JustPressed() {
		SpawnBlessing(player.actor.Position(), player.actor.FacingVec(), player.id.Handle)
	}

	// Launch editor
	if settings.ActionLaunchEditor.JustPressed() {
		cmd := exec.Command("./Total Editor 3.exe", gWorld.GameMap.Name)
		err := cmd.Run()
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 100 {
			gWorld.app.ProcessSignal(game.MapChangeSignal{
				NextMapPath:     gWorld.GameMap.Name,
				GiveAmmo:        player.ammo,
				GiveArmor:       player.armorType,
				ArmorAmount:     player.armorAmount,
				EquippedWeapons: player.equippedWeapons,
			})
		} else if err != nil {
			failure.LogErrWithLocation("failed to launch editor: %v", err)
		}
	}

	// Spawn chicken
	if settings.ActionSpawnChicken.JustPressed() {
		SpawnChicken(player.Body().Position.Add(player.actor.FacingVec()), mgl32.Vec3{})
	}

	// Use key
	if settings.Current.ActionUse.JustPressed() {
		const useDist float32 = 3.0
		hit, closestBody := gWorld.Raycast(player.Body().Position, player.actor.FacingVec(), ColLayerUsable, useDist, player.Body())
		if hit.Hit && !closestBody.IsNil() {
			if usable, isUsable := scene.Get[Usable](closestBody); isUsable {
				usable.OnUse(player)
			}
		}
	} else {
		// Use items by walking into them
		ents := gWorld.bspTree.PotentiallyTouchingEnts(player.Body().Position, player.Body().Shape)
		for handle := range ents {
			item, ok := scene.Get[*Item](handle)
			if !ok || item.Body() == nil || !item.Body().OnLayer(ColLayerUsable) {
				continue
			}
			if item.Body().Shape.Touches(item.Body().Position, player.Body().Position.Add(player.Body().Velocity.Mul(deltaTime)), player.Body().Shape) {
				item.OnUse(player)
			}
		}
	}

	// Weapon selection
	switch true {
	case settings.Current.ActionSickle.JustPressed():
		player.selectedWeapon = game.WeaponSickle
	case settings.Current.ActionChicken.JustPressed():
		player.selectedWeapon = game.WeaponChicken
	case settings.Current.ActionGrenade.JustPressed():
		player.selectedWeapon = game.WeaponGrenade
	case settings.Current.ActionParusu.JustPressed():
		player.selectedWeapon = game.WeaponParusu
	case settings.Current.ActionDblGrenade.JustPressed():
		player.selectedWeapon = game.WeaponDblGrenade
	case settings.Current.ActionSign.JustPressed():
		player.selectedWeapon = game.WeaponSign
	case settings.Current.ActionAirhorn.JustPressed():
		player.selectedWeapon = game.WeaponAirhorn
	case settings.Current.ActionDefenestrator.JustPressed():
		player.selectedWeapon = game.WeaponDefenestrator
	case settings.Current.ActionCluckster.JustPressed():
		player.selectedWeapon = game.WeaponCluckster
	}

	// Fire weapon
	if player.safety {
		if !settings.Current.ActionFire.Pressed() {
			player.safety = false
		}
	} else if player.SelectedWeapon != nil && settings.Current.ActionFire.Pressed() {
		var cast collision.Result
		if player.SelectedWeapon.IsShooter {
			// Don't fire if there is a wall too close in front
			cast, _ = gWorld.Raycast(player.Body().Position, player.actor.FacingVec(), ColLayerMap, 1.5, player.Body())
		}

		ammoBefore := player.ammo
		//TODO: Where these damn parameters ccomin from dawg?
		if !cast.Hit && player.SelectedWeapon.AttemptFire(&player.ammo, deltaTime) {
			player.punTimer.Reset()
			player.actor.noisyTimer = 0.5

			if player.armorType == game.ArmorTypeBullet && player.SelectedWeapon != &player.Sickle {
				player.ammo = ammoBefore
			}
		}
	}

	// Sprinting
	if settings.Current.ActionSlow.Pressed() {
		player.actor.MaxSpeed = player.WalkSpeed
	} else {
		player.actor.MaxSpeed = player.RunSpeed
	}

	if settings.Current.ActionWeaponWheel.Pressed() {
		player.weaponWheelOpenness = 1.0
	}
	if player.weaponWheelOpenness <= 0.0 {
		sensitivity := float32(0.005 * math2.Pow(10.0, (settings.Current.MouseSensitivity-1.0)/5.0))
		normalLook := settings.Current.ActionLookLeft.Axis() - settings.Current.ActionLookRight.Axis()
		fastLook := (settings.Current.ActionFastLookLeft.Axis() - settings.Current.ActionFastLookRight.Axis()) * 2.5
		player.actor.YawAngle += (normalLook + fastLook) * sensitivity
	}
}

func (player *Player) ProcessSignal(signal any) {
	switch signal.(type) {
	case game.TeleportationSignal:
		gWorld.Hud.FlashScreen(color.Color{R: 1.0, G: 0.0, B: 1.0, A: 1.0}, 1.0)
	case game.ResumeGameSignal:
		player.safety = true
	}
}

func (player *Player) OnDamage(sourceEntity any, damage float32) bool {
	if player.godMode {
		return false
	}

	hudPtr := &gWorld.Hud

	wasNonZero := player.armorAmount > 0
	player.armorAmount = max(0, player.armorAmount-damage)
	if player.armorAmount <= 0 && wasNonZero {
		player.armorType = game.ArmorTypeNone
		cache.GetSfx("assets/sounds/armor_break.wav").Play()
		hudPtr.ShowMessage(settings.Localize("armorBroken"), 10, color.Red)
	}
	damage *= (1.0 - player.armorType.Defense())

	player.actor.Health = max(0, player.actor.Health-damage)

	if player.actor.Health > 0 {
		hudPtr.FlashScreen(color.Red.WithAlpha(0.5), 0.75)

		if bodyHaver, ok := sourceEntity.(comps.HasBody); ok {
			// Change the hurt face with respect to the direction the damage is coming from
			dmgDir := bodyHaver.Body().Position.Sub(player.Body().Position)
			if dmgDir.LenSqr() > 0.0 {
				dmgDir = dmgDir.Normalize()
			}
			forward := player.actor.FacingVec()
			halfFov := mgl32.DegToRad(float32(settings.Current.Fov) / 2.0)
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
