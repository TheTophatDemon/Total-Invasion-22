package world

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type ActorsInSphereIter struct {
	ActorsIter
	radiusSq  float32
	spherePos mgl32.Vec3
	exception HasActor
}

func (iter *ActorsInSphereIter) Next() (HasActor, scene.Handle) {
	for {
		actorEnt, actorId := iter.ActorsIter.Next()
		if actorEnt == nil {
			break
		}
		if actorEnt == iter.exception {
			continue
		}
		body := actorEnt.Body()
		if body.Transform.Position().Sub(iter.spherePos).LenSqr() < iter.radiusSq {
			return actorEnt, actorId
		}
	}
	return nil, scene.Handle{}
}

func (world *World) IterActorsInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception HasActor) ActorsInSphereIter {
	return ActorsInSphereIter{
		ActorsIter: world.IterActors(),
		radiusSq:   sphereRadius * sphereRadius,
		spherePos:  spherePos,
		exception:  exception,
	}
}

// TODO: Replace with iterator.
func (world *World) BodiesInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception comps.HasBody) []scene.Handle {
	result := make([]scene.Handle, 0)
	iter := world.IterBodies()
	for bodyEnt, bodyId := iter.Next(); bodyEnt != nil; bodyEnt, bodyId = iter.Next() {
		if bodyEnt == exception {
			continue
		}
		body := bodyEnt.Body()

		if collision.NewSphere(sphereRadius).Touches(spherePos, body.Transform.Position(), body.Shape) {
			result = append(result, bodyId)
		}
	}
	return result
}

func (world *World) AnyProjectilesInSphere(spherePos mgl32.Vec3, sphereRadius float32) bool {
	iter := world.Projectiles.Iter()
	for proj, _ := iter.Next(); proj != nil; proj, _ = iter.Next() {
		body := proj.Body()
		if collision.NewSphere(sphereRadius).Touches(spherePos, body.Transform.Position(), body.Shape) {
			return true
		}
	}
	return false
}

func (world *World) Raycast(rayOrigin, rayDir mgl32.Vec3, filter collision.Mask, maxDist float32, excludeBody comps.HasBody) (collision.RaycastResult, scene.Handle) {
	var rayBB math2.Box = math2.BoxFromPoints(rayOrigin, rayOrigin.Add(rayDir.Mul(maxDist)))
	var closestEnt scene.Handle
	var closestBodyHit collision.RaycastResult
	closestBodyHit.Distance = math.MaxFloat32
	iter := world.IterBodies()
	for bodyEnt, bodyId := iter.Next(); bodyEnt != nil; bodyEnt, bodyId = iter.Next() {
		body := bodyEnt.Body()
		if bodyEnt == excludeBody ||
			!bodyEnt.Body().OnLayer(filter) ||
			!body.Shape.Extents().Translate(body.Transform.Position()).Intersects(rayBB) {
			continue
		}
		bodyHit := body.Shape.Raycast(rayOrigin, rayDir, body.Transform.Position(), maxDist)
		if bodyHit.Hit && bodyHit.Distance < closestBodyHit.Distance {
			closestBodyHit = bodyHit
			closestEnt = bodyId
		}
	}
	if !closestEnt.IsNil() {
		return closestBodyHit, closestEnt
	}
	return collision.RaycastResult{}, scene.Handle{}
}

// Returns an iterator over all linkables with the given non-zero link number.
func (world *World) NextLinkableWithNumber(iter *LinkablesIter, linkNumber int) (Linkable, scene.Handle) {
	if linkNumber == 0 {
		return nil, scene.Handle{}
	}

	for ent, id := iter.Next(); ent != nil; ent, id = iter.Next() {
		if ent.LinkNumber() == linkNumber {
			return ent, id
		}
	}
	return nil, scene.Handle{}
}

func (world *World) ActivateLinks(source Linkable) {
	iter := world.IterLinkables()
	for {
		ent, handle := world.NextLinkableWithNumber(&iter, source.LinkNumber())
		if ent == nil {
			break
		}
		if !handle.Equals(source.Handle()) {
			ent.OnLinkActivate(source)
		}
	}
}

func (world *World) DeactivateLinks(source Linkable) {
	iter := world.IterLinkables()
	for {
		ent, handle := world.NextLinkableWithNumber(&iter, source.LinkNumber())
		if ent == nil {
			break
		}
		if !handle.Equals(source.Handle()) {
			ent.OnLinkDeactivate(source)
		}
	}
}
