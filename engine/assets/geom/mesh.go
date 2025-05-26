package geom

import (
	"fmt"
	"log"
	"math"
	"slices"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Group struct {
	Offset, Length int
}

type Mesh struct {
	verts Vertices
	inds  []uint32

	tris          []math2.Triangle // Mathematical representation of the triangles (will be empty until Triangles() is called)
	bbox          math2.Box        // Bounding box over vertices. Lazily evaluated.
	groups        map[string]Group
	primitiveType uint32

	uploaded              bool
	vertBuffer, idxBuffer uint32 //VBOs
	vertArray             uint32 //VAO
}

type MeshTriangleIter struct {
	mesh          *Mesh
	triangleIndex int
}

func CreateMesh(verts Vertices, inds []uint32) *Mesh {
	mesh := &Mesh{
		verts:         verts,
		inds:          inds,
		tris:          nil,
		groups:        make(map[string]Group, 0),
		primitiveType: gl.TRIANGLES,
	}
	return mesh
}

func CreateWireMesh(verts Vertices, inds []uint32) *Mesh {
	return &Mesh{
		verts:         verts,
		inds:          inds,
		tris:          nil,
		groups:        make(map[string]Group, 0),
		primitiveType: gl.LINES,
	}
}

func WireMeshFromTriangleMesh(triangleMesh *Mesh, col color.Color) *Mesh {
	wireVerts := Vertices{
		Pos:   slices.Clone(triangleMesh.Verts().Pos),
		Color: make([]mgl32.Vec4, len(triangleMesh.Verts().Pos)),
	}

	for i := range wireVerts.Color {
		wireVerts.Color[i] = col.Vector()
	}

	wireInds := make([]uint32, 0, len(triangleMesh.inds)*2)
	for i := 0; i < len(triangleMesh.inds); i += 3 {
		wireInds = append(wireInds, triangleMesh.inds[i], triangleMesh.inds[i+1])
		wireInds = append(wireInds, triangleMesh.inds[i], triangleMesh.inds[i+2])
		wireInds = append(wireInds, triangleMesh.inds[i+1], triangleMesh.inds[i+2])
	}

	return CreateWireMesh(wireVerts, wireInds)
}

func WireMeshFromBoundingBox(bbox math2.Box) *Mesh {
	corners := bbox.Corners()
	boxVerts := Vertices{
		Pos: corners[:],
		Color: []mgl32.Vec4{
			{1.0, 0.0, 1.0, 1.0},
			{1.0, 0.0, 1.0, 1.0},
			{1.0, 0.0, 1.0, 1.0},
			{1.0, 0.0, 1.0, 1.0},
			{1.0, 0.0, 1.0, 1.0},
			{1.0, 0.0, 1.0, 1.0},
			{1.0, 0.0, 1.0, 1.0},
			{1.0, 0.0, 1.0, 1.0},
		},
	}
	return CreateWireMesh(boxVerts, []uint32{
		math2.CORNER_BLT, math2.CORNER_BRT,
		math2.CORNER_BLT, math2.CORNER_BLB,
		math2.CORNER_BRT, math2.CORNER_BRB,
		math2.CORNER_BRB, math2.CORNER_BLB,
		math2.CORNER_FLT, math2.CORNER_FRT,
		math2.CORNER_FLT, math2.CORNER_FLB,
		math2.CORNER_FRT, math2.CORNER_FRB,
		math2.CORNER_FRB, math2.CORNER_FLB,
		math2.CORNER_FLT, math2.CORNER_BLT,
		math2.CORNER_FRT, math2.CORNER_BRT,
		math2.CORNER_FLB, math2.CORNER_BLB,
		math2.CORNER_FRB, math2.CORNER_BRB,
	})
}

func (m *Mesh) SetGroup(name string, group Group) {
	m.groups[name] = group
}

func (m *Mesh) HasGroup(name string) bool {
	_, ok := m.groups[name]
	return ok
}

func (m *Mesh) Group(name string) Group {
	return m.groups[name]
}

func (m *Mesh) GroupCount() int {
	return len(m.groups)
}

func (m *Mesh) GroupNames() []string {
	out := make([]string, 0, len(m.groups))
	for name := range m.groups {
		out = append(out, name)
	}
	return out
}

func (m *Mesh) Verts() Vertices {
	return m.verts
}

func (m *Mesh) Inds() []uint32 {
	return m.inds
}

// Iterates over the triangles in the mesh (triangles are lazily evaluated)
func (mesh *Mesh) IterTriangles() MeshTriangleIter {
	if mesh.tris == nil && mesh.primitiveType == gl.TRIANGLES {
		// Determine the triangles from the indices & vertex positions.
		mesh.tris = make([]math2.Triangle, len(mesh.inds)/3)
		for t := range mesh.tris {
			mesh.tris[t] = math2.Triangle{
				mesh.verts.Pos[mesh.inds[t*3+0]],
				mesh.verts.Pos[mesh.inds[t*3+1]],
				mesh.verts.Pos[mesh.inds[t*3+2]],
			}
		}
	}

	return MeshTriangleIter{
		mesh:          mesh,
		triangleIndex: 0,
	}
}

func (iter *MeshTriangleIter) Next() math2.Triangle {
	if iter == nil || iter.mesh == nil {
		return math2.Triangle{}
	}
	tri := iter.mesh.tris[iter.triangleIndex]
	iter.triangleIndex += 1
	return tri
}

func (iter *MeshTriangleIter) HasNext() bool {
	return iter != nil && iter.mesh != nil && iter.triangleIndex < len(iter.mesh.tris)
}

func (iter *MeshTriangleIter) Collect() []math2.Triangle {
	if iter == nil || iter.mesh == nil {
		return []math2.Triangle{}
	}
	tris := make([]math2.Triangle, 0, len(iter.mesh.tris))
	for iter.HasNext() {
		tris = append(tris, iter.Next())
	}
	return tris
}

// Returns the total number of triangles in the mesh.
func (iter *MeshTriangleIter) Count() int {
	if iter == nil || iter.mesh == nil {
		return 0
	}
	return len(iter.mesh.tris)
}

func (m *Mesh) BoundingBox() math2.Box {
	if m.bbox.Size().LenSqr() < mgl32.Epsilon {
		// Calculate the bounding box if it hasn't been calculated already.
		m.bbox.Max = mgl32.Vec3{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
		m.bbox.Min = mgl32.Vec3{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
		for _, vert := range m.verts.Pos {
			m.bbox.Min = math2.Vec3Min(m.bbox.Min, vert)
			m.bbox.Max = math2.Vec3Max(m.bbox.Max, vert)
		}
	}
	return m.bbox
}

// Calculates this mesh's axis aligned bounding box if its vertices were transformed by the given matrix.
func (m *Mesh) TransformedAABB(transform mgl32.Mat4) math2.Box {
	var bbox math2.Box
	bbox.Max = mgl32.Vec3{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
	bbox.Min = mgl32.Vec3{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
	for _, vert := range m.verts.Pos {
		vert = mgl32.TransformCoordinate(vert, transform)
		bbox.Min = math2.Vec3Min(bbox.Min, vert)
		bbox.Max = math2.Vec3Max(bbox.Max, vert)
	}
	return bbox
}

func (m *Mesh) Bind() {
	if m == nil {
		return
	}
	if !m.uploaded {
		m.Upload()
	}

	gl.BindVertexArray(m.vertArray)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vertBuffer)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.idxBuffer)
	m.verts.BindAttributes()
}

func (m *Mesh) Upload() {
	if m.uploaded {
		m.Free()
	}

	//Flatten the vertex array into a series of floats
	data, err := m.verts.Flatten()
	if err != nil {
		log.Println("Error: Invalid vertex data for mesh upload.")
		return
	}

	//Create VAO
	gl.GenVertexArrays(1, &m.vertArray)
	gl.BindVertexArray(m.vertArray)

	//Create buffers

	//Vertex buffer
	gl.GenBuffers(1, &m.vertBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vertBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(data)*int(unsafe.Sizeof(data[0])), gl.Ptr(data), gl.STATIC_DRAW)

	//Index buffer
	gl.GenBuffers(1, &m.idxBuffer)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.idxBuffer)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER,
		len(m.inds)*int(unsafe.Sizeof(m.inds[0])), //Size in bytes of buffer data
		gl.Ptr(m.inds),
		gl.STATIC_DRAW)

	m.uploaded = true
}

func (m *Mesh) DrawAll() {
	if m == nil {
		return
	}
	gl.DrawElementsWithOffset(m.primitiveType, int32(len(m.inds)), gl.UNSIGNED_INT, 0)
}

func (m *Mesh) DrawGroup(name string) error {
	group, ok := m.groups[name]
	if !ok {
		return fmt.Errorf("Group not found")
	}
	gl.DrawElementsWithOffset(m.primitiveType, int32(group.Length), gl.UNSIGNED_INT, uintptr(group.Offset)*unsafe.Sizeof(m.inds[0]))
	return nil
}

func (m *Mesh) Free() {
	gl.DeleteBuffers(1, &m.vertBuffer)
	gl.DeleteBuffers(1, &m.idxBuffer)
	gl.DeleteVertexArrays(1, &m.vertArray)
}
