package collision

import (
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type ShapeKind uint8

const (
	SHAPE_KIND_NONE ShapeKind = iota
	SHAPE_KIND_SPHERE
	SHAPE_KIND_BOX
	SHAPE_KIND_MESH
)

func (sk ShapeKind) String() string {
	switch sk {
	case SHAPE_KIND_NONE:
		return "None"
	case SHAPE_KIND_SPHERE:
		return "Sphere"
	case SHAPE_KIND_BOX:
		return "Box"
	case SHAPE_KIND_MESH:
		return "Mesh"
	}
	return "Invalid"
}

type Shape struct {
	kind    ShapeKind
	radius  float32      // The radius of the sphere collision shape.
	extents math2.Box    // Describes the body's bounding box, centered at its origin.
	mesh    *assets.Mesh // Refers to the mesh used for the mesh collision shape.
}

func (s *Shape) String() string {
	return s.kind.String()
}

func (s *Shape) Extents() math2.Box {
	return s.extents
}

func (s *Shape) Kind() ShapeKind {
	return s.kind
}

func ShapeSphere(radius float32) Shape {
	return Shape{
		kind:    SHAPE_KIND_SPHERE,
		radius:  radius,
		extents: math2.BoxFromRadius(radius),
	}
}

// Returns the shape's radius and 'true' if the shape is a sphere.
func (s *Shape) Radius() (float32, bool) {
	if s.kind != SHAPE_KIND_SPHERE {
		return 0.0, false
	}
	return s.radius, true
}

func ShapeBox(extents math2.Box) Shape {
	return Shape{
		kind:    SHAPE_KIND_BOX,
		extents: extents,
	}
}

func ShapeMesh(mesh *assets.Mesh) Shape {
	if mesh == nil {
		panic("mesh must not be nil")
	}
	return Shape{
		kind:    SHAPE_KIND_MESH,
		extents: mesh.BoundingBox(),
		mesh:    mesh,
	}
}

// Returns the shape's mesh and 'true' if the shape is a mesh.
func (s *Shape) Mesh() (*assets.Mesh, bool) {
	if s.kind != SHAPE_KIND_MESH {
		return nil, false
	}
	return s.mesh, true
}
