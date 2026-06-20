package world

import (
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/failure"
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
	Yaw             float32
	id              scene.Id[*Trigger]
	particles       comps.ParticleRender
	filter          func(comps.HasBody) bool
	onEnter         func(trigger *Trigger, entHandle scene.Handle)
	whileTouching   func(trigger *Trigger, entHandle scene.Handle, deltaTime float32)
	onExit          func(trigger *Trigger, entHandle scene.Handle)
	linkNumber      int
	touching        [triggerMaxContacts]scene.Handle
	damagePerSecond float32
	entProperties   map[string]string // Properties on the te3 entity used to spawn this trigger.
}

var _ Linkable = (*Trigger)(nil)

func SpawnTriggerFromTE3(ent te3.Ent) (id scene.Id[*Trigger], tr *Trigger, err error) {
	id, tr, err = gWorld.Triggers.New()
	if err != nil {
		return
	}

	tr.id = id
	tr.Radius = ent.Radius
	tr.Position = ent.Position
	trans := comps.TransformFromTE3Ent(ent, false, false)
	tr.Yaw = trans.Yaw()
	tr.linkNumber, _ = ent.IntProperty("link")
	tr.entProperties = ent.Properties

	switch ent.Properties["action"] {
	case TriggerActionTeleport:
		tr.filter = liveActorsOnlyFilter
		tr.onEnter = teleportAction
		tr.particles = TeleportParticles(0.5)
		tr.particles.Init()
	case TriggerActionDamage:
		tr.filter = liveActorsOnlyFilter
		tr.whileTouching = damageWhileTouching
		damageRate, err := strconv.ParseFloat(ent.Properties["damagePerSecond"], 32)
		if err != nil || math.IsNaN(damageRate) {
			damageRate = 0.0
		}
		tr.damagePerSecond = float32(damageRate)
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
		bodyHaver, _ := scene.Get[comps.HasBody](handle)
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
	teleportingEnt, _ := scene.Get[HasActor](handle)
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
					victimEnt, _ := scene.Get[HasActor](actorHandle)
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

	gWorld.EnterWinState("assets/maps/"+tr.entProperties["level"]+".te3", cameraHandle)
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
	timeStr := tr.entProperties["messageTime"]
	if timeStr != "" {
		log.Println("Warning: 'messageTime' property for triggers is obsolete")
	}

	priorityStr := tr.entProperties["messagePriority"]
	priority, err := strconv.ParseInt(priorityStr, 10, 32)
	if err != nil {
		if len(priorityStr) != 0 {
			failure.LogErrWithLocation("invalid message priority specified: %v", priorityStr)
		}
		priority = 10
	}

	colr := color.Color{A: 1.0}
	colrStrs := strings.Split(tr.entProperties["messageColor"], ",")
	if len(colrStrs) == 3 {
		for i, str := range colrStrs {
			val, err := strconv.ParseInt(str, 10, 32)
			if err != nil {
				colr = color.Color{}
				break
			}
			floatVal := float32(val) / 255.0
			switch i {
			case 0:
				colr.R = floatVal
			case 1:
				colr.G = floatVal
			case 2:
				colr.B = floatVal
			}
		}
	}
	if colr == (color.Color{}) {
		if len(colrStrs) != 0 {
			failure.LogErrWithLocation("invalid message color specified: %v", colrStrs)
		}
		colr = color.White
	}

	gWorld.Hud.ShowMessage(settings.Localize(tr.entProperties["messageKey"]), int(priority), colr)
}

func damageWhileTouching(tr *Trigger, handle scene.Handle, deltaTime float32) {
	if damageable, canDamage := scene.Get[Damageable](handle); canDamage {
		damageable.OnDamage(tr, tr.damagePerSecond*deltaTime)
	}
}

func checkpointAction(tr *Trigger, handle scene.Handle) {
	gWorld.app.ProcessSignal(game.SaveSignal{
		Number: 0,
	})
	tr.onEnter = nil // Prevent duplicate saves
	gWorld.Hud.FlashScreen(color.Green, 0.5)
	gWorld.Hud.ShowMessage(settings.Localize("checkpoint"), 20, color.Green)
	cache.GetSfx("assets/sounds/checkpoint.wav").Play()
}

func liveActorsOnlyFilter(ent comps.HasBody) bool {
	actorHaver, ok := ent.(HasActor)
	if !ok {
		return false
	}
	return actorHaver.Actor().Health > 0
}

func playerOnlyFilter(ent comps.HasBody) bool {
	_, isPlayer := ent.(*Player)
	return isPlayer
}
