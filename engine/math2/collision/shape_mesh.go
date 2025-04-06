package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Mesh struct {
	shape
	triangles []math2.Triangle
}

var _ Shape = (*Mesh)(nil)

func NewMesh(mesh *geom.Mesh, scale float32) Mesh {
	if mesh == nil {
		panic("mesh must not be nil")
	}
	triIter := mesh.IterTriangles()
	triangles := triIter.Collect()
	if scale != 1.0 {
		// Scale the triangles
		for i := range triangles {
			triangles[i][0] = triangles[i][0].Mul(scale)
			triangles[i][1] = triangles[i][1].Mul(scale)
			triangles[i][2] = triangles[i][2].Mul(scale)
		}
	}
	return Mesh{
		shape: shape{
			extents: mesh.BoundingBox(),
		},
		triangles: triangles,
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

func (mesh Mesh) String() string {
	return "Mesh"
}

func (mesh Mesh) Triangles() []math2.Triangle {
	return mesh.triangles
}

func (mesh Mesh) Extents() math2.Box {
	return mesh.extents
}

func (mesh Mesh) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3, maxDist float32) (cast RaycastResult) {
	var nearestHitDist float32 = math.MaxFloat32
	for _, triangle := range mesh.triangles {
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

func (mesh Mesh) ResolveCollision(myPosition, myMovement, theirPosition mgl32.Vec3, theirShape Shape) Result {
	panic("collision resolution not implemented for mesh")
}

func WireMeshFromMeshCollisionShape(meshShape *Mesh, col color.Color) *geom.Mesh {
	tris := meshShape.Triangles()

	wireVerts := geom.Vertices{
		Pos:   make([]mgl32.Vec3, len(tris)*3),
		Color: make([]mgl32.Vec4, len(tris)*3),
	}
	wireInds := make([]uint32, 0, len(tris)*3)

	for _, tri := range tris {
		baseIndex := uint32(len(wireVerts.Pos))
		for v := range tri {
			wireVerts.Pos = append(wireVerts.Pos, tri[v])
			wireVerts.Color = append(wireVerts.Color, col.Vector())
		}
		wireInds = append(wireInds, baseIndex, baseIndex+1, baseIndex, baseIndex+2, baseIndex+1, baseIndex+2)
	}

	return geom.CreateWireMesh(wireVerts, wireInds)
}
