package collision

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Shape interface {
	Extents() math2.Box                                              // Returns the body's bounding box, centered at its origin.
	Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3) RaycastResult // Test the shape for collision against a ray
	ResolveCollision(myPosition, theirPosition mgl32.Vec3, theirShape Shape) Result
}

type shape struct {
	extents math2.Box
}

type Box struct {
	shape
}

var _ Shape = (*Box)(nil)

func NewBox(extents math2.Box) Box {
	return Box{
		shape: shape{
			extents: extents,
		},
	}
}

func (b Box) Extents() math2.Box {
	return b.extents
}

func (b Box) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3) RaycastResult {
	return RayBoxCollision(rayOrigin, rayDir, b.extents.Translate(shapeOffset))
}

func (b Box) ResolveCollision(myPosition, theirPosition mgl32.Vec3, theirShape Shape) Result {
	panic("collision resolution not implemented for boxes")
}

type Sphere struct {
	shape
	radius float32
}

var _ Shape = (*Sphere)(nil)

func NewSphere(radius float32) Sphere {
	return Sphere{
		shape: shape{
			extents: math2.BoxFromRadius(radius),
		},
		radius: radius,
	}
}

func (s Sphere) String() string {
	return "Sphere"
}

func (s Sphere) Extents() math2.Box {
	return s.extents
}

func (s Sphere) Radius() float32 {
	return s.radius
}

func (s Sphere) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3) RaycastResult {
	return RaySphereCollision(rayOrigin, rayDir, shapeOffset, s.radius)
}

func (s Sphere) ResolveCollision(myPosition, theirPosition mgl32.Vec3, theirShape Shape) Result {
	switch otherShape := theirShape.(type) {
	case Sphere:
		return ResolveSphereSphere(myPosition, theirPosition, s, otherShape)
	case Box:
		return ResolveSphereBox(myPosition, theirPosition, s, otherShape)
	case Mesh:
		return ResolveSphereTriangles(myPosition, theirPosition, s, otherShape, otherShape.triangleIndices, TRI_PART_ALL)
	}
	return Result{}
}

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

func (m Mesh) Raycast(rayOrigin, rayDir, shapeOffset mgl32.Vec3) (cast RaycastResult) {
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
		if newCast.Hit {
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
