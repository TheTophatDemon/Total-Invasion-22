package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Mesh struct {
	shape
	mesh            *geom.Mesh
	triangleIndices []int
}

var _ Shape = (*Mesh)(nil)

func NewMesh(mesh *geom.Mesh) Mesh {
	if mesh == nil {
		panic("mesh must not be nil")
	}
	return Mesh{
		shape: shape{
			extents: mesh.BoundingBox(),
		},
		mesh:            mesh,
		triangleIndices: nil,
	}
}

func NewMeshSubset(mesh *geom.Mesh, triangleIndices []int) Mesh {
	if mesh == nil {
		panic("mesh must not be nil")
	}
	return Mesh{
		shape: shape{
			extents: mesh.BoundingBox(),
		},
		mesh:            mesh,
		triangleIndices: triangleIndices,
	}
}

func (m Mesh) Mesh() *geom.Mesh {
	return m.mesh
}

func (m Mesh) Extents() math2.Box {
	return m.mesh.BoundingBox()
}

func (m Mesh) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) (cast RaycastResult) {
	meshTris := m.mesh.Triangles()
	loopLimit := len(meshTris)
	if m.triangleIndices != nil {
		loopLimit = len(m.triangleIndices)
	}
	var nearestHitDist float32 = math.MaxFloat32
	for i := 0; i < loopLimit; i++ {
		var triangle math2.Triangle
		if m.triangleIndices == nil {
			triangle = meshTris[i]
		} else {
			triangle = meshTris[m.triangleIndices[i]]
		}

		newCast := RayTriangleCollision(rayOrigin, rayDir, triangle)
		if newCast.Hit && newCast.Distance <= maxDist {
			if dist := newCast.Position.Sub(rayOrigin).LenSqr(); dist < nearestHitDist {
				nearestHitDist = dist
				cast = newCast
			}
		}
	}
	return
}

func (m Mesh) ResolveCollision(myPosition, theirPosition mgl32.Vec3, theirShape Shape) Result {
	panic("collision resolution not implemented for mesh")
}
