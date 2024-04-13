package world

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

// Iterates through all of the stores in the world that can be cast to the given Component type.
func iterateStores[Component any](world *World) func() (Component, scene.Handle) {
	var zero Component
	iters := make([]func() (any, scene.Handle), 0, 10)
	scene.ForEachStorageField(world, func(storage scene.StorageOps) {
		iters = append(iters, storage.IterUntyped())
	})
	var iter int = 0
	return func() (Component, scene.Handle) {
		for ; iter < len(iters); iter++ {
			if item, id := iters[iter](); item != nil {
				if component, ok := item.(Component); ok {
					return component, id
				}
			}
		}
		return zero, scene.Handle{}
	}
}

func (world *World) BodiesIter() func() (comps.HasBody, scene.Handle) {
	return iterateStores[comps.HasBody](world)
}

func (world *World) ActorsIter() func() (HasActor, scene.Handle) {
	return iterateStores[HasActor](world)
}

func (world *World) LinkablesIter(linkNumber int) func() (Linkable, scene.Handle) {
	return iterateStores[Linkable](world)
}

func (world *World) ActorsInSphere(spherePos mgl32.Vec3, sphereRadius float32, exception HasActor) []scene.Handle {
	radiusSq := sphereRadius * sphereRadius
	nextActor := world.ActorsIter()
	result := make([]scene.Handle, 0)
	for actorEnt, actorId := nextActor(); actorEnt != nil; actorEnt, actorId = nextActor() {
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
	nextBody := world.BodiesIter()
	result := make([]scene.Handle, 0)
	for bodyEnt, bodyId := nextBody(); bodyEnt != nil; bodyEnt, bodyId = nextBody() {
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

func (world *World) Raycast(rayOrigin, rayDir mgl32.Vec3, filter collision.Mask, maxDist float32, excludeBody comps.HasBody) (collision.RaycastResult, scene.Handle) {
	var rayBB math2.Box = math2.BoxFromPoints(rayOrigin, rayOrigin.Add(rayDir.Mul(maxDist)))
	var closestEnt scene.Handle
	var closestBodyHit collision.RaycastResult
	var nextBody func() (comps.HasBody, scene.Handle) = world.BodiesIter()
	closestBodyHit.Distance = math.MaxFloat32
	for bodyEnt, bodyId := nextBody(); bodyEnt != nil; bodyEnt, bodyId = nextBody() {
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
