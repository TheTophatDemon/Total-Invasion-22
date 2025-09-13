package world

import (
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
	TriggerActionTeleport = "teleport"
	TriggerActionDamage   = "damage"
	TriggerActionEndLevel = "end level"
	TriggerActionSecret   = "secret"
	TriggerActionActivate = "activate"
	TriggerActionMessage  = "message"
)

type Trigger struct {
	Radius          float32
	Transform       comps.Transform
	id              scene.Id[*Trigger]
	particles       comps.ParticleRender
	filter          func(comps.HasBody) bool
	onEnter         func(trigger *Trigger, entHandle scene.Handle)
	whileTouching   func(trigger *Trigger, entHandle scene.Handle, deltaTime float32)
	onExit          func(trigger *Trigger, entHandle scene.Handle)
	world           *World
	linkNumber      int
	touching        [triggerMaxContacts]scene.Handle
	damagePerSecond float32
	entProperties   map[string]string // Properties on the te3 entity used to spawn this trigger.
}

var _ Linkable = (*Trigger)(nil)

func SpawnTriggerFromTE3(world *World, ent te3.Ent) (id scene.Id[*Trigger], tr *Trigger, err error) {
	id, tr, err = world.Triggers.New()
	if err != nil {
		return
	}

	tr.world = world
	tr.id = id
	tr.Radius = ent.Radius
	tr.Transform = comps.TransformFromTE3Ent(ent, false, false)
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
		world.Hud.VictoryScreen.SecretsTotal++
	case TriggerActionActivate:
		tr.filter = playerOnlyFilter
		tr.onEnter = activateAction
	case TriggerActionMessage:
		tr.filter = playerOnlyFilter
		tr.onEnter = messageAction
	}

	return
}

func (tr *Trigger) Update(deltaTime float32) {
	// Call callbacks for new & already touching entities
	touchingNow := tr.world.IterBodiesInSphere(tr.Transform.Position(), tr.Radius, nil)
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

	tr.particles.Update(deltaTime, &tr.Transform)
}

func (tr *Trigger) Render(context *render.Context) {
	tr.particles.Render(&tr.Transform, context)
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
	iter := tr.world.IterLinkables()
	for {
		link, _ := tr.world.NextLinkableWithNumber(&iter, tr.linkNumber)
		if link == nil {
			break
		}
		if link != tr {
			if trOther, isTrigger := link.(*Trigger); isTrigger {
				// If there are NPCs standing on the other side, kill them.
				actorsIter := tr.world.IterActorsInSphere(trOther.Transform.Position(), trOther.Radius, nil)
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

				teleportingBody.Transform.SetPosition(trOther.Transform.Position())
				teleportingBody.Velocity = mgl32.Vec3{}
				actor := teleportingEnt.Actor()
				actor.SetYaw(trOther.Transform.Yaw())
				actor.inputForward, actor.inputStrafe = 0.0, 0.0
				teleportingEnt.ProcessSignal(game.TeleportationSignal{})
				// This registers with the other teleporter that the body is touching without triggering the onEnter() callback,
				// which would cause the destination teleporter to immediately teleport the body back.
				trOther.addToTouching(handle)
				const sfxTeleport = "assets/sounds/teleport.wav"
				cache.GetSfx(sfxTeleport).PlayAttenuatedV(actor.Position())
				cache.GetSfx(sfxTeleport).PlayAttenuatedV(tr.Transform.Position())
				tr.particles.EmissionTimer = 0.5
				trOther.particles.EmissionTimer = 0.5

				break
			}
		}
	}
}

func exitLevelAction(tr *Trigger, handle scene.Handle) {
	var cameraHandle scene.Handle
	iter := tr.world.IterLinkables()
	for {
		linkable, id := tr.world.NextLinkableWithNumber(&iter, tr.linkNumber)
		if linkable == nil {
			break
		}
		if _, isCamera := linkable.(*Camera); isCamera {
			cameraHandle = id
			break
		}
	}
	if cameraHandle.IsNil() {
		cameraHandle = tr.world.CurrentCamera.Handle
	}

	tr.world.EnterWinState("assets/maps/"+tr.entProperties["level"]+".te3", cameraHandle)
}

func secretAreaAction(tr *Trigger, handle scene.Handle) {
	tr.world.Hud.VictoryScreen.SecretsFound++
	tr.world.Hud.ShowMessage(settings.Localize("foundSecret"), 2.0, 50, color.Red)
	cache.GetSfx("assets/sounds/secret_chime.wav").Play()
	tr.world.QueueRemoval(tr.id.Handle)
}

func activateAction(tr *Trigger, handle scene.Handle) {
	tr.world.ActivateLinks(tr)
}

func messageAction(tr *Trigger, handle scene.Handle) {
	timeStr := tr.entProperties["messageTime"]
	time, err := strconv.ParseFloat(timeStr, 32)
	if err != nil {
		if len(timeStr) != 0 {
			failure.LogErrWithLocation("invalid message time specified: %v", timeStr)
		}
		time = 1.0
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

	tr.world.Hud.ShowMessage(settings.Localize(tr.entProperties["messageKey"]), float32(time), int(priority), colr)
}

func damageWhileTouching(tr *Trigger, handle scene.Handle, deltaTime float32) {
	if damageable, canDamage := scene.Get[Damageable](handle); canDamage {
		damageable.OnDamage(tr, tr.damagePerSecond*deltaTime)
	}
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
