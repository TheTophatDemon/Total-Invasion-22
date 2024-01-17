package ents

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

const TRIGGER_TOUCH_MAX = 3

type Trigger struct {
	Sphere        collision.Sphere
	Transform     comps.Transform
	filter        func(comps.HasBody) bool
	onEnter       func(*Trigger, scene.Handle)
	whileTouching func(*Trigger, scene.Handle)
	onExit        func(*Trigger, scene.Handle)
	world         WorldOps
	linkNumber    int
	touching      [TRIGGER_TOUCH_MAX]scene.Handle
}

var _ Linkable = (*Trigger)(nil)

func NewTriggerFromTE3(ent te3.Ent, world WorldOps) (tr Trigger, err error) {
	tr.world = world
	tr.Sphere = collision.NewSphere(ent.Radius)
	tr.Transform = ent.Transform(false, false)
	tr.linkNumber, _ = ent.IntProperty("link")

	switch ent.Properties["action"] {
	case "teleport":
		tr.filter = actorsOnlyFilter
		tr.onEnter = teleportAction
	}

	return
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
					tr.whileTouching(tr, handle)
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
		if !tr.touching[i].IsNil() || !tr.touching[i].Exists() {
			tr.touching[i] = handle
			return true, i
		}
	}
	return false, -1
}

func teleportAction(tr *Trigger, handle scene.Handle) {
	links := tr.world.LinkablesIter(tr.linkNumber)
	for link, _ := links(); link != nil; link, _ = links() {
		if link != tr && link.LinkNumber() == tr.linkNumber {
			if trOther, isTrigger := link.(*Trigger); isTrigger {
				bodyHaver, _ := scene.Get[comps.HasBody](handle)
				bodyHaver.Body().Transform.SetPosition(trOther.Transform.Position())
				// This registers with the other teleporter that the body is touching without triggering the onEnter() callback,
				// which would cause the destination teleporter to immediately teleport the body back.
				trOther.addToTouching(handle)
				break
			}
		}
	}
}

func actorsOnlyFilter(e comps.HasBody) bool {
	_, ok := e.(HasActor)
	return ok
}
