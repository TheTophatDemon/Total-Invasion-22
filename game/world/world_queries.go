package world

import (
	"iter"
	"log"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

func (world *World) ActorsInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception HasActor) []scene.Handle {
	radiusSq := sphereRadius * sphereRadius
	result := make([]scene.Handle, 0)
	for actorId, actorEnt := range world.AllActors() {
		if actorEnt == exception {
			continue
		}
		body := actorEnt.Body()
		if body.Transform.Position().Sub(spherePos).LenSqr() < radiusSq {
			result = append(result, actorId)
		}
	}
	return result
}

func (world *World) BodiesInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception comps.HasBody) []scene.Handle {
	result := make([]scene.Handle, 0)
	for bodyId, bodyEnt := range world.AllBodies() {
		if bodyEnt == exception {
			continue
		}
		body := bodyEnt.Body()

		var hit bool
		switch shape := body.Shape.(type) {
		case collision.Sphere:
			hit = collision.SphereTouchesSphere(spherePos, sphereRadius, body.Transform.Position(), shape.Radius())
		case collision.Box:
			hit = collision.SphereTouchesBox(spherePos, sphereRadius, shape.Extents().Translate(body.Transform.Position()))
		case collision.Mesh:
			for _, tri := range shape.Triangles() {
				if h, _ := collision.SphereTriangleCollision(spherePos, sphereRadius, tri, body.Transform.Position()); h != collision.TRI_PART_NONE {
					hit = true
					break
				}
			}
		}
		if hit {
			result = append(result, bodyId)
		}
	}
	return result
}

func (world *World) ProjectilesInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception *Projectile) []scene.Handle {
	result := make([]scene.Handle, 0)
	for projId, proj := range world.Projectiles.All() {
		if proj == exception {
			continue
		}
		body := proj.Body()

		var hit bool
		switch shape := body.Shape.(type) {
		case collision.Sphere:
			hit = collision.SphereTouchesSphere(spherePos, sphereRadius, body.Transform.Position(), shape.Radius())
		default:
			log.Printf("Warning: invalid collision shape for projectile detected (%s)\n", shape.String())
		}
		if hit {
			result = append(result, projId)
		}
	}
	return result
}

func (world *World) Raycast(rayOrigin, rayDir mgl32.Vec3, filter collision.Mask, maxDist float32, excludeBody comps.HasBody) (collision.RaycastResult, scene.Handle) {
	var rayBB math2.Box = math2.BoxFromPoints(rayOrigin, rayOrigin.Add(rayDir.Mul(maxDist)))
	var closestEnt scene.Handle
	var closestBodyHit collision.RaycastResult
	closestBodyHit.Distance = math.MaxFloat32
	for bodyId, bodyEnt := range world.AllBodies() {
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
func (world *World) LinkablesWithNumber(linkNumber int) iter.Seq2[scene.Handle, Linkable] {
	return func(yield func(scene.Handle, Linkable) bool) {
		// Link number 0 is the nil value
		if linkNumber == 0 {
			return
		}
		nextLinkable, stop := iter.Pull2(world.AllLinkables())
		defer stop()
		for handle, ent, ok := nextLinkable(); ok; handle, ent, ok = nextLinkable() {
			if ent.LinkNumber() == linkNumber && !yield(handle, ent) {
				return
			}
		}
	}
}

func (world *World) ActivateLinks(source Linkable) {
	for handle, ent := range world.LinkablesWithNumber(source.LinkNumber()) {
		if !handle.Equals(source.Handle()) {
			ent.OnLinkActivate(source)
		}
	}
}

func (world *World) DeactivateLinks(source Linkable) {
	for handle, ent := range world.LinkablesWithNumber(source.LinkNumber()) {
		if !handle.Equals(source.Handle()) {
			ent.OnLinkDeactivate(source)
		}
	}
}
