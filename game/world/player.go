package world

import (
	"iter"
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

	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	Player struct {
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
		SelectedWeapon, NextWeapon                                                            *Weapon
		Sickle, Chicken, Grenade, Parusu, DblGrenade, Sign, Airhorn, Defenestrator, Cluckster Weapon
		WeaponWheel                                                                           HudWeaponWheel
		punTimer                                                                              timer.Timer
		puns                                                                                  []string
		safety                                                                                bool // Prevents firing accidentally after exiting a menu
	}
)

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
	player.ammo = changeInfo.Equipment.Ammo
	player.ammo[game.AmmoTypeSickle] = 0
	player.armorType = changeInfo.Equipment.Armor
	player.armorAmount = changeInfo.Equipment.ArmorAmount

	// Initialize weapons
	player.Sickle.Init(&WeaponSickle, true)
	player.SelectedWeapon = &player.Sickle
	player.Sickle.State = WeaponStateIntro
	player.Chicken.Init(&WeaponChicken, changeInfo.Equipment.EquippedWeapons[game.WeaponIndexChicken])
	player.Grenade.Init(&WeaponGrenade, changeInfo.Equipment.EquippedWeapons[game.WeaponIndexGrenade])
	player.Parusu.Init(&WeaponParusu, changeInfo.Equipment.EquippedWeapons[game.WeaponIndexParusu])
	player.DblGrenade.Init(&WeaponDblGrenade, changeInfo.Equipment.EquippedWeapons[game.WeaponIndexDblGrenade])
	player.Sign.Init(&WeaponSign, changeInfo.Equipment.EquippedWeapons[game.WeaponIndexSign])
	player.Airhorn.Init(&WeaponAirhorn, changeInfo.Equipment.EquippedWeapons[game.WeaponIndexAirhorn])
	player.Defenestrator.Init(&WeaponDefenestrator, changeInfo.Equipment.EquippedWeapons[game.WeaponIndexDefenestrator])
	player.Cluckster.Init(&WeaponCluckster, changeInfo.Equipment.EquippedWeapons[game.WeaponIndexCluckster])

	if gWorld.Hud.Intro.TimeLeft() > 0.0 {
		// Spawn intro sickle
		firePos := player.Body().Position.Add(player.actor.FacingVec().Mul(80.0))
		SpawnIntroSickle(firePos, player.actor.FacingVec(), player.id.Handle)
	} else {
		player.ammo[game.AmmoTypeSickle] = 1
	}

	return
}

func (player *Player) Equipment() game.Equipment {
	equipment := game.Equipment{
		Ammo:        player.ammo,
		Armor:       player.armorType,
		ArmorAmount: player.armorAmount,
		EquippedWeapons: [...]bool{
			game.WeaponIndexSickle:        player.Sickle.Equipped,
			game.WeaponIndexChicken:       player.Chicken.Equipped,
			game.WeaponIndexGrenade:       player.Grenade.Equipped,
			game.WeaponIndexParusu:        player.Parusu.Equipped,
			game.WeaponIndexDblGrenade:    player.DblGrenade.Equipped,
			game.WeaponIndexSign:          player.Sign.Equipped,
			game.WeaponIndexAirhorn:       player.Airhorn.Equipped,
			game.WeaponIndexDefenestrator: player.Defenestrator.Equipped,
			game.WeaponIndexCluckster:     player.Cluckster.Equipped,
		},
		Keys: player.keys,
	}
	if player.SelectedWeapon != nil {
		equipment.SelectedWeapon = player.SelectedWeapon.Index
	}
	return equipment
}

func (player *Player) Update(deltaTime float32) {
	hudPtr := &gWorld.Hud

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
				NextMapPath: gWorld.impendingLevel,
				Equipment:   player.Equipment(),
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
		player.TrySelect(nil)
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

	// Transition selected weapon
	if player.SelectedWeapon == nil || player.SelectedWeapon.State == WeaponStateInactive {
		player.SelectedWeapon = player.NextWeapon
		player.SelectedWeapon.OnSelect()
	}

	// Update weapon wheel
	if !settings.Current.ActionWeaponWheel.Pressed() {
		player.WeaponWheel.Openness = max(0.0, player.WeaponWheel.Openness-(deltaTime*10.0))
	}
	if settings.Current.ActionWeaponWheel.JustPressed() {
		player.WeaponWheel = NewWeaponWheel(player)
	} else if settings.Current.ActionWeaponWheel.JustReleased() && player.actor.Health > 0 {
		weap := player.WeaponWithIndex(player.WeaponWheel.HighlightedWeapon)
		player.TrySelect(weap)
	}

	// Update weapons
	for weap := range player.Weapons() {
		weap.Update(player, deltaTime)
	}

	player.actor.Update(deltaTime)
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
		for weap := range player.Weapons() {
			weap.Equipped = true
		}
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
				NextMapPath: gWorld.GameMap.Name,
				Equipment:   player.Equipment(),
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
		player.TrySelect(&player.Sickle)
	case settings.Current.ActionChicken.JustPressed():
		player.TrySelect(&player.Chicken)
	case settings.Current.ActionGrenade.JustPressed():
		player.TrySelect(&player.Grenade)
	case settings.Current.ActionParusu.JustPressed():
		player.TrySelect(&player.Parusu)
	case settings.Current.ActionDblGrenade.JustPressed():
		player.TrySelect(&player.DblGrenade)
	case settings.Current.ActionSign.JustPressed():
		player.TrySelect(&player.Sign)
	case settings.Current.ActionAirhorn.JustPressed():
		player.TrySelect(&player.Airhorn)
	case settings.Current.ActionDefenestrator.JustPressed():
		player.TrySelect(&player.Defenestrator)
	case settings.Current.ActionCluckster.JustPressed():
		player.TrySelect(&player.Cluckster)
	}

	// Sprinting
	if settings.Current.ActionSlow.Pressed() {
		player.actor.MaxSpeed = player.WalkSpeed
	} else {
		player.actor.MaxSpeed = player.RunSpeed
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
		if !cast.Hit && player.SelectedWeapon.AttemptFire(player, deltaTime, settings.Current.ActionFire.JustPressed()) {
			player.punTimer.Reset()
			player.actor.NoiseLevel = 0.5

			if player.armorType == game.ArmorTypeBullet && player.SelectedWeapon != &player.Sickle {
				player.ammo = ammoBefore
			}
		} else if player.SelectedWeapon == &player.Sign && player.Sign.CooldownTimer > 0.0 {
			// Slow down the player when swinging the sign
			player.actor.MaxSpeed = player.WalkSpeed
		}
	}

	if settings.Current.ActionWeaponWheel.Pressed() {
		player.WeaponWheel.Openness = 1.0
	}
	if player.WeaponWheel.Openness <= 0.0 {
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
				hudPtr.StatusBar.SuggestPlayerFace(FaceStateHurtFront)
			} else if forward.Cross(dmgDir).Y() > 0.0 {
				// Source is to the left
				hudPtr.StatusBar.SuggestPlayerFace(FaceStateHurtLeft)
			} else {
				// Source is to the right
				hudPtr.StatusBar.SuggestPlayerFace(FaceStateHurtRight)
			}
		} else {
			hudPtr.StatusBar.SuggestPlayerFace(FaceStateHurtFront)
		}
	}
	return true
}

func (player *Player) WeaponWithIndex(index game.WeaponIndex) *Weapon {
	for weap := range player.Weapons() {
		if weap.Index == index {
			return weap
		}
	}
	return nil
}

func (player *Player) Weapons() iter.Seq[*Weapon] {
	return func(yield func(*Weapon) bool) {
		for _, weap := range [...]*Weapon{
			&player.Sickle,
			&player.Chicken,
			&player.Grenade,
			&player.Parusu,
			&player.DblGrenade,
			&player.Sign,
			&player.Airhorn,
			&player.Defenestrator,
			&player.Cluckster,
		} {
			if !yield(weap) {
				return
			}
		}
	}
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

func (player *Player) TrySelect(weapon *Weapon) {
	if player.SelectedWeapon == weapon {
		return
	}
	if weapon != nil && !weapon.Equipped {
		gWorld.Hud.ShowMessage(settings.Localize(weapon.Name+"NotFound"), 30, color.Red)
		return
	}
	if player.SelectedWeapon != nil {
		player.SelectedWeapon.OnDeselect()
	}
	player.NextWeapon = weapon
}
