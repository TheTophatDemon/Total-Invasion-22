package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const triggerMaxContacts = 3

const (
	TriggerActionTeleport   = "teleport"
	TriggerActionDamage     = "damage"
	TriggerActionEndLevel   = "end level"
	TriggerActionSecret     = "secret"
	TriggerActionActivate   = "activate"
	TriggerActionMessage    = "message"
	TriggerActionCheckpoint = "checkpoint"
)

type Trigger struct {
	Radius          float32
	Position        mgl32.Vec3
	Yaw             math2.Radians
	id              scene.Id[*Trigger]
	particles       comps.ParticleRender
	filter          func(comps.HasBody) bool
	onEnter         func(trigger *Trigger, entHandle scene.Handle)
	whileTouching   func(trigger *Trigger, entHandle scene.Handle, deltaTime float32)
	onExit          func(trigger *Trigger, entHandle scene.Handle)
	linkNumber      int
	touching        [triggerMaxContacts]scene.Handle
	damagePerSecond float32
	entProperties   game.EntProps // Properties on the te3 entity used to spawn this trigger.
}

var _ Linkable = (*Trigger)(nil)

func SpawnTriggerFromTE3(ent game.EntDef) (id scene.Id[*Trigger], tr *Trigger, err error) {
	id, tr, err = gWorld.Triggers.New()
	if err != nil {
		return
	}

	tr.id = id
	tr.Radius = ent.Radius
	tr.Position = ent.Position
	trans := comps.TransformFromTE3Ent(ent.Ent, false, false)
	tr.Yaw = math2.Radians(trans.Yaw())
	tr.linkNumber, _ = ent.Properties.Link.Value()
	tr.entProperties = ent.Properties

	switch ent.Properties.Action.Or("") {
	case TriggerActionTeleport:
		tr.filter = liveActorsOnlyFilter
		tr.onEnter = teleportAction
		tr.particles = TeleportParticles(0.5)
		tr.particles.Init()
	case TriggerActionDamage:
		tr.filter = liveActorsOnlyFilter
		tr.whileTouching = damageWhileTouching
		tr.damagePerSecond = ent.Properties.DamagePerSecond.Or(0.0)
	case TriggerActionEndLevel:
		tr.filter = playerOnlyFilter
		tr.onEnter = exitLevelAction
	case TriggerActionSecret:
		tr.filter = playerOnlyFilter
		tr.onEnter = secretAreaAction
		gWorld.Hud.VictoryScreen.SecretsTotal++
	case TriggerActionActivate:
		tr.filter = playerOnlyFilter
		tr.onEnter = activateAction
	case TriggerActionMessage:
		tr.filter = playerOnlyFilter
		tr.onEnter = messageAction
	case TriggerActionCheckpoint:
		tr.filter = playerOnlyFilter
		tr.onEnter = checkpointAction
	}

	return
}

func (tr *Trigger) Update(deltaTime float32) {
	// Call callbacks for new & already touching entities
	touchingNow := gWorld.IterBodiesInSphere(tr.Position, tr.Radius, nil)
	var stillTouching [triggerMaxContacts]bool
	for _, handle := touchingNow.Next(); !handle.IsNil(); _, handle = touchingNow.Next() {
		bodyHaver, _ := handle.Get[comps.HasBody]()
		if tr.filter == nil || tr.filter(bodyHaver) {
			if added, index := tr.addToTouching(handle); added {
				if tr.onEnter != nil {
					tr.onEnter(tr, handle)
				}
				stillTouching[index] = true
			} else if index >= 0 {
				if tr.whileTouching != nil {
					tr.whileTouching(tr, handle, deltaTime)
				}
				stillTouching[index] = true
			}
		}
	}
	// Remove entities no longer being touched
	for i := range stillTouching {
		if !stillTouching[i] && !tr.touching[i].IsNil() {
			if tr.onExit != nil && tr.touching[i].Exists() {
				tr.onExit(tr, tr.touching[i])
			}
			tr.touching[i] = scene.Handle{}
		}
	}

	tr.particles.Update(deltaTime, tr.Position)
}

func (tr *Trigger) Render(context *render.Context) {
	tr.particles.Render(tr.Position, context)
}

func (tr *Trigger) LinkNumber() int {
	return tr.linkNumber
}

func (tr *Trigger) Handle() scene.Handle {
	return tr.id.Handle
}

func (tr *Trigger) OnLinkActivate(source Linkable) {
}

func (tr *Trigger) OnLinkDeactivate(source Linkable) {
}

// Returns a bool that is true if the handle was added to a new slot.
// The int returned is the index of the handle in the array if found, or -1.
func (tr *Trigger) addToTouching(handle scene.Handle) (bool, int) {
	for i := range tr.touching {
		if !tr.touching[i].IsNil() && tr.touching[i].Equals(handle) {
			return false, i
		}
	}
	for i := range tr.touching {
		if tr.touching[i].IsNil() || !tr.touching[i].Exists() {
			tr.touching[i] = handle
			return true, i
		}
	}
	return false, -1
}

func teleportAction(tr *Trigger, handle scene.Handle) {
	teleportingEnt, _ := handle.Get[HasActor]()
	if teleportingEnt.Actor().Health <= 0 {
		return
	}
	teleportingBody := teleportingEnt.Body()
	iter := gWorld.IterLinkables()
	for {
		link, _ := gWorld.NextLinkableWithNumber(&iter, tr.linkNumber)
		if link == nil {
			break
		}
		if link != tr {
			if trOther, isTrigger := link.(*Trigger); isTrigger {
				// If there are NPCs standing on the other side, kill them.
				actorsIter := gWorld.IterActorsInSphere(trOther.Position, trOther.Radius, nil)
				for {
					_, actorHandle := actorsIter.Next()
					if actorHandle.IsNil() {
						break
					}
					victimEnt, _ := actorHandle.Get[HasActor]()
					if player, isPlayer := victimEnt.(*Player); isPlayer && player != teleportingEnt {
						// If the player is on the other side, kill the NPC instead.
						teleportingEnt.(Damageable).OnDamage(tr, math2.Inf32())
						return
					} else if teleportingEnt == victimEnt {
						continue
					}
					victimEnt.(Damageable).OnDamage(tr, math2.Inf32())
				}

				teleportingBody.Position = trOther.Position
				teleportingBody.Velocity = mgl32.Vec3{}
				actor := teleportingEnt.Actor()
				actor.SetYaw(trOther.Yaw)
				actor.inputForward, actor.inputStrafe = 0.0, 0.0
				teleportingEnt.ProcessSignal(game.TeleportationSignal{})
				// This registers with the other teleporter that the body is touching without triggering the onEnter() callback,
				// which would cause the destination teleporter to immediately teleport the body back.
				trOther.addToTouching(handle)
				const sfxTeleport = "assets/sounds/teleport.wav"
				cache.GetSfx(sfxTeleport).PlayAttenuatedV(actor.Position())
				cache.GetSfx(sfxTeleport).PlayAttenuatedV(tr.Position)
				tr.particles.EmissionTimer = 0.5
				trOther.particles.EmissionTimer = 0.5

				break
			}
		}
	}
}

func exitLevelAction(tr *Trigger, handle scene.Handle) {
	var cameraHandle scene.Handle
	iter := gWorld.IterLinkables()
	for {
		linkable, id := gWorld.NextLinkableWithNumber(&iter, tr.linkNumber)
		if linkable == nil {
			break
		}
		if _, isCamera := linkable.(*Camera); isCamera {
			cameraHandle = id
			break
		}
	}
	if cameraHandle.IsNil() {
		cameraHandle = gWorld.CurrentCamera.Handle
	}

	gWorld.EnterWinState("assets/maps/"+tr.entProperties.Level.Or("")+".te3", cameraHandle)
}

func secretAreaAction(tr *Trigger, handle scene.Handle) {
	gWorld.Hud.VictoryScreen.SecretsFound++
	gWorld.Hud.ShowMessage(settings.Localize("foundSecret"), 50, color.Red)
	cache.GetSfx("assets/sounds/secret_chime.wav").Play()
	gWorld.QueueRemoval(tr.id.Handle)
}

func activateAction(tr *Trigger, handle scene.Handle) {
	gWorld.ActivateLinks(tr)
}

func messageAction(tr *Trigger, handle scene.Handle) {
	priority := tr.entProperties.MessagePriority.Or(10)

	colr := color.Color{A: 1.0}
	colorVec := tr.entProperties.MessageColor.Or(mgl32.Vec3{255, 255, 255})
	colr.R = colorVec[0] / 255.0
	colr.G = colorVec[1] / 255.0
	colr.B = colorVec[2] / 255.0

	gWorld.Hud.ShowMessage(settings.Localize(tr.entProperties.MessageKey.Or("")), priority, colr)
}

func damageWhileTouching(tr *Trigger, handle scene.Handle, deltaTime float32) {
	if damageable, canDamage := handle.Get[Damageable](); canDamage {
		damageable.OnDamage(tr, tr.damagePerSecond*deltaTime)
	}
}

func checkpointAction(tr *Trigger, handle scene.Handle) {
	tr.id.Remove() // Prevent duplicate saves
	gWorld.Hud.FlashScreen(color.Green, 0.5)
	gWorld.Hud.ShowMessage(settings.Localize("checkpoint"), 20, color.Green)
	cache.GetSfx("assets/sounds/checkpoint.wav").Play()
	gWorld.ProcessSignal(game.SaveSignal{
		Number:          0,
		AfterCheckpoint: true,
	})
}

func liveActorsOnlyFilter(ent comps.HasBody) bool {
	actorHaver, ok := ent.(HasActor)
	if !ok {
		return false
	}
	return actorHaver.Actor().Health > 0
}

func playerOnlyFilter(ent comps.HasBody) bool {
	player, isPlayer := ent.(*Player)
	return isPlayer && player.Actor().Health > 0
}

func (trigger *Trigger) Save() game.EntDef {
	return game.EntDef{
		Position:   trigger.Position,
		Radius:     trigger.Radius,
		Display:    te3.ENT_DISPLAY_SPHERE,
		Color:      [3]int{255, 0, 255},
		Properties: trigger.entProperties,
	}
}
