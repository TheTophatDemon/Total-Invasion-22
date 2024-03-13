package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Mesh struct {
	shape
	triangles []math2.Triangle
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
		triangles: mesh.Triangles(),
	}
}

func NewMeshFromTriangles(triangles []math2.Triangle) Mesh {
	if triangles == nil {
		panic("triangles must not be nil")
	}
	points := make([]mgl32.Vec3, len(triangles)*3)
	for i, tri := range triangles {
		points[i*3] = tri[0]
		points[i*3+1] = tri[1]
		points[i*3+2] = tri[2]
	}
	return Mesh{
		shape: shape{
			extents: math2.BoxFromPoints(points...),
		},
		triangles: triangles,
	}
}

func (m Mesh) Triangles() []math2.Triangle {
	return m.triangles
}

func (m Mesh) Extents() math2.Box {
	return m.extents
}

func (m Mesh) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) (cast RaycastResult) {
	var nearestHitDist float32 = math.MaxFloat32
	for _, triangle := range m.triangles {
		triangle = triangle.OffsetBy(shapeOffset)
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
