package world

import (
	"fmt"
	"iter"
	"math"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
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

func SpawnPlayerFromTE3(
	ent te3.Ent,
	camera scene.Id[*Camera],
) (id scene.Id[*Player], player *Player, err error) {
	id, player, err = gWorld.Players.New()
	if err != nil {
		return
	}
	player.id = id
	player.actor = Actor{
		body: comps.Body{
			Position: ent.Position,
			Shape:    collision.NewBoxShape(0.6, 0.7, 0.6),
			Layers:   ColLayerActors | ColLayerPlayers,
		},
		collisionFilter: ColLayerMap | ColLayerActors | ColLayerInvisible,
		YawAngle:        math2.ToRadians(math2.Degrees(ent.Angles[1])),
		AccelRate:       100.0,
		Friction:        20.0,
		MaxHealth:       200,
		TargetHealth:    100,
		Health:          ent.FloatPropertyOr("health", 100),
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
	player.ammo[game.AmmoTypeSickle] = 0
	player.ammo[game.AmmoTypeEgg] = ent.IntPropertyOr("ammoEgg", 0)
	player.ammo[game.AmmoTypeGrenade] = ent.IntPropertyOr("ammoGrenade", 0)
	player.ammo[game.AmmoTypePlasma] = ent.IntPropertyOr("ammoPlasma", 0)
	switch ent.Properties["armor"] {
	case "boring":
		player.armorType = game.ArmorTypeBoring
	case "bullet":
		player.armorType = game.ArmorTypeBullet
	case "super":
		player.armorType = game.ArmorTypeSuper
	case "chronos":
		player.armorType = game.ArmorTypeChronos
	}
	player.armorAmount = ent.FloatPropertyOr("armorAmount", 0.0)

	player.keys = game.Keys(ent.IntPropertyOr("keys", 0))

	// Initialize weapons
	player.Sickle.Init(&WeaponSickle, true)
	player.SelectedWeapon = &player.Sickle
	player.Sickle.State = WeaponStateIntro
	chickenEquipped := ent.BoolPropertyOr("chickenEquipped", false)
	player.Chicken.Init(&WeaponChicken, chickenEquipped)
	grenadeEquipped := ent.BoolPropertyOr("grenadeEquipped", false)
	player.Grenade.Init(&WeaponGrenade, grenadeEquipped)
	parusuEquipped := ent.BoolPropertyOr("parusuEquipped", false)
	player.Parusu.Init(&WeaponParusu, parusuEquipped)
	dblGrenadeEquipped := ent.BoolPropertyOr("dblGrenadeEquipped", false)
	player.DblGrenade.Init(&WeaponDblGrenade, dblGrenadeEquipped)
	signEquipped := ent.BoolPropertyOr("signEquipped", false)
	player.Sign.Init(&WeaponSign, signEquipped)
	airhornEquipped := ent.BoolPropertyOr("airhornEquipped", false)
	player.Airhorn.Init(&WeaponAirhorn, airhornEquipped)
	defenestratorEquipped := ent.BoolPropertyOr("defenestratorEquipped", false)
	player.Defenestrator.Init(&WeaponDefenestrator, defenestratorEquipped)
	clucksterEquipped := ent.BoolPropertyOr("clucksterEquipped", false)
	player.Cluckster.Init(&WeaponCluckster, clucksterEquipped)

	if gWorld.Hud.Intro.TimeLeft() > 0.0 {
		// Spawn intro sickle
		firePos := player.Body().Position.Add(player.actor.FacingVec().Mul(80.0))
		SpawnIntroSickle(firePos, player.actor.FacingVec(), player.id.Handle)
	} else {
		player.ammo[game.AmmoTypeSickle] = 1
	}

	return
}

func (player *Player) Save() te3.Ent {
	ent := te3.Ent{
		Angles:   [3]math2.Degrees{0, math2.ToDegrees(player.actor.YawAngle), 0},
		Position: player.actor.Position(),
		Texture:  "assets/textures/sprites/segan.png",
		Radius:   0.7,
		Display:  te3.ENT_DISPLAY_SPRITE,
		Color:    [3]uint8{255, 255, 255},
		Properties: map[string]string{
			"type":                  "player",
			"ammoEgg":               fmt.Sprintf("%d", player.ammo[game.AmmoTypeEgg]),
			"ammoGrenade":           fmt.Sprintf("%d", player.ammo[game.AmmoTypeGrenade]),
			"ammoPlasma":            fmt.Sprintf("%d", player.ammo[game.AmmoTypePlasma]),
			"armor":                 player.armorType.Name(),
			"armorAmount":           fmt.Sprintf("%.2f", player.armorAmount),
			"chickenEquipped":       fmt.Sprintf("%t", player.Chicken.Equipped),
			"health":                fmt.Sprintf("%.2f", player.actor.Health),
			"grenadeEquipped":       fmt.Sprintf("%t", player.Grenade.Equipped),
			"parusuEquipped":        fmt.Sprintf("%t", player.Parusu.Equipped),
			"dblGrenadeEquipped":    fmt.Sprintf("%t", player.DblGrenade.Equipped),
			"signEquipped":          fmt.Sprintf("%t", player.Sign.Equipped),
			"airhornEquipped":       fmt.Sprintf("%t", player.Airhorn.Equipped),
			"defenestratorEquipped": fmt.Sprintf("%t", player.Defenestrator.Equipped),
			"clucksterEquipped":     fmt.Sprintf("%t", player.Cluckster.Equipped),
			"keys":                  fmt.Sprintf("%d", player.keys),
		},
	}
	return ent
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
				MapPath:   gWorld.impendingLevel,
				PlayerEnt: new(player.Save()),
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
			gWorld.app.ProcessSignal(game.MapChangeSignal{MapPath: gWorld.GameMap.Name})
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
				MapPath:   gWorld.GameMap.Name,
				PlayerEnt: new(player.Save()),
			})
		} else if err != nil {
			failure.LogErrWithLocation("failed to launch editor: %v", err)
		}
	}

	// Spawn chicken
	if settings.ActionSpawnChicken.JustPressed() {
		SpawnChicken(player.Body().Position.Add(player.actor.FacingVec()), [3]math2.Degrees{})
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
		player.actor.YawAngle += math2.Radians((normalLook + fastLook) * sensitivity)
	}
}

func (player *Player) ProcessSignal(signal any) {
	switch signal.(type) {
	case game.TeleportationSignal:
		gWorld.Hud.FlashScreen(color.Color{R: 1.0, G: 0.0, B: 1.0, A: 1.0}, 1.0)
	case game.ResumeGameSignal:
		player.safety = true
		// Reinitialize weapon sprites for potential new screen resolution
		player.NextWeapon = player.SelectedWeapon
		for weap := range player.Weapons() {
			weap.Init(weap.WeaponDef, weap.Equipped)
		}
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
