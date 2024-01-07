package ents

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Trigger struct {
	Sphere     collision.Sphere
	Transform  comps.Transform
	filter     func(comps.HasBody) bool
	onEnter    func(*Trigger, comps.HasBody)
	world      WorldOps
	linkNumber int
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
	if tr.onEnter == nil {
		return
	}

	intersectionists := tr.world.BodiesInSphere(tr.Transform.Position(), tr.Sphere.Radius(), nil)
	for _, bodyHaver := range intersectionists {
		if tr.filter == nil || tr.filter(bodyHaver) {
			// TODO: Track bodies that have already entered
			tr.onEnter(tr, bodyHaver)
		}
	}
}

func (tr *Trigger) LinkNumber() int {
	return tr.linkNumber
}

func teleportAction(tr *Trigger, e comps.HasBody) {
	links := tr.world.LinkablesIter(tr.linkNumber)
	for link := links(); link != nil; link = links() {
		if link != tr {
			if bodyHaver, hasBody := link.(comps.HasBody); hasBody {
				e.Body().Transform.SetPosition(bodyHaver.Body().Transform.Position())
			}
			break
		}
	}
}

func actorsOnlyFilter(e comps.HasBody) bool {
	_, ok := e.(HasActor)
	return ok
}
