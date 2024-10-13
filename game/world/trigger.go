package world

import (
	"fmt"
	"math"
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/world/effects"
)

const (
	TRIGGER_TOUCH_MAX = 3
	SFX_TELEPORT      = "assets/sounds/teleport.wav"
)

const (
	TRIGGER_ACTION_TELEPORT  = "teleport"
	TRIGGER_ACTION_DAMAGE    = "damage"
	TRIGGER_ACTION_END_LEVEL = "end level"
	TRIGGER_ACTION_SECRET    = "secret"
)

const (
	TRIGGER_ACTION      = "action"
	TRIGGER_DAMAGE_RATE = "damagePerSecond"
)

type Trigger struct {
	Sphere          collision.Sphere
	Transform       comps.Transform
	id              scene.Id[*Trigger]
	particles       comps.ParticleRender
	filter          func(comps.HasBody) bool
	onEnter         func(trigger *Trigger, entHandle scene.Handle)
	whileTouching   func(trigger *Trigger, entHandle scene.Handle, deltaTime float32)
	onExit          func(trigger *Trigger, entHandle scene.Handle)
	world           *World
	linkNumber      int
	touching        [TRIGGER_TOUCH_MAX]scene.Handle
	damagePerSecond float32
	nextLevel       string
}

var _ Linkable = (*Trigger)(nil)

func SpawnTriggerFromTE3(world *World, ent te3.Ent) (id scene.Id[*Trigger], tr *Trigger, err error) {
	id, tr, err = world.Triggers.New()
	if err != nil {
		return
	}

	tr.world = world
	tr.id = id
	tr.Sphere = collision.NewSphere(ent.Radius)
	tr.Transform = comps.TransformFromTE3Ent(ent, false, false)
	tr.linkNumber, _ = ent.IntProperty("link")

	switch ent.Properties[TRIGGER_ACTION] {
	case TRIGGER_ACTION_TELEPORT:
		tr.filter = liveActorsOnlyFilter
		tr.onEnter = teleportAction
		tr.particles = effects.Teleport(0.5)
		tr.particles.Init()
	case TRIGGER_ACTION_DAMAGE:
		tr.filter = liveActorsOnlyFilter
		tr.whileTouching = damageWhileTouching
		damageRate, err := strconv.ParseFloat(ent.Properties[TRIGGER_DAMAGE_RATE], 32)
		if err != nil || math.IsNaN(damageRate) {
			damageRate = 0.0
		}
		tr.damagePerSecond = float32(damageRate)
	case TRIGGER_ACTION_END_LEVEL:
		tr.filter = playerOnlyFilter
		tr.onEnter = exitLevelAction
		tr.nextLevel = ent.Properties["level"]
	case TRIGGER_ACTION_SECRET:
		tr.filter = playerOnlyFilter
		tr.onEnter = secretAreaAction
		world.Hud.SecretsTotal++
	}

	return
}

func SpawnKillzone(world *World, position mgl32.Vec3, radius float32, damagePerSecond float32) (id scene.Id[*Trigger], tr *Trigger, err error) {
	return SpawnTriggerFromTE3(world, te3.Ent{
		Radius:   radius,
		Position: position,
		Properties: map[string]string{
			TRIGGER_ACTION:      TRIGGER_ACTION_DAMAGE,
			TRIGGER_DAMAGE_RATE: fmt.Sprintf("%f", damagePerSecond),
		},
	})
}

func (tr *Trigger) Update(deltaTime float32) {
	// Call callbacks for new & already touching entities
	touchingNow := tr.world.BodiesInSphere(tr.Transform.Position(), tr.Sphere.Radius(), nil)
	var stillTouching [TRIGGER_TOUCH_MAX]bool
	for _, handle := range touchingNow {
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
	for _, link := range tr.world.LinkablesWithNumber(tr.linkNumber) {
		if link != tr {
			if trOther, isTrigger := link.(*Trigger); isTrigger {
				// If there are NPCs standing on the other side, kill them.
				actors := tr.world.ActorsInSphere(trOther.Transform.Position(), trOther.Sphere.Radius(), nil)
				for _, actorHandle := range actors {
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
				cache.GetSfx(SFX_TELEPORT).PlayAttenuatedV(actor.Position())
				cache.GetSfx(SFX_TELEPORT).PlayAttenuatedV(tr.Transform.Position())
				tr.particles.EmissionTimer = 0.5
				trOther.particles.EmissionTimer = 0.5

				break
			}
		}
	}
}

func exitLevelAction(tr *Trigger, handle scene.Handle) {
	var cameraHandle scene.Handle
	for id, linkable := range tr.world.LinkablesWithNumber(tr.linkNumber) {
		if _, isCamera := linkable.(*Camera); isCamera {
			cameraHandle = id
			break
		}
	}
	if cameraHandle.IsNil() {
		cameraHandle = tr.world.CurrentCamera.Handle
	}

	tr.world.EnterWinState(tr.nextLevel, cameraHandle)
}

func secretAreaAction(tr *Trigger, handle scene.Handle) {
	tr.world.Hud.SecretsFound++
	tr.world.QueueRemoval(tr.id.Handle)
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
