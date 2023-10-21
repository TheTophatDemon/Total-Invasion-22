package geom

import (
	"fmt"
	"log"
	"math"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

const (
	SIZEOF_POS      = 3 * 4
	SIZEOF_TEXCOORD = 2 * 4
	SIZEOF_NORMAL   = 3 * 4
	SIZEOF_COLOR    = 4 * 4
)

const (
	ATTR_POS = iota
	ATTR_TEXCOORD
	ATTR_NORMAL
	ATTR_COLOR
)

type Vertices struct {
	Pos      []mgl32.Vec3
	TexCoord []mgl32.Vec2
	Normal   []mgl32.Vec3
	Color    []mgl32.Vec4
}

func (v *Vertices) Stride() int {
	stride := 0
	if v.Pos != nil && len(v.Pos) > 0 {
		stride += SIZEOF_POS
	}
	if v.TexCoord != nil && len(v.TexCoord) > 0 {
		stride += SIZEOF_TEXCOORD
	}
	if v.Normal != nil && len(v.Normal) > 0 {
		stride += SIZEOF_NORMAL
	}
	if v.Color != nil && len(v.Color) > 0 {
		stride += SIZEOF_COLOR
	}
	return stride
}

func (verts *Vertices) Flatten() ([]float32, error) {
	if len(verts.Pos) <= 0 {
		return nil, fmt.Errorf("vertices has no position data")
	}
	data := make([]float32, 0, len(verts.Pos)*verts.Stride())
	for v, pos := range verts.Pos {
		data = append(data, pos.X(), pos.Y(), pos.Z())
		if verts.TexCoord != nil && v < len(verts.TexCoord) {
			data = append(data, verts.TexCoord[v].X(), verts.TexCoord[v].Y())
		}
		if verts.Normal != nil && v < len(verts.Normal) {
			data = append(data, verts.Normal[v].X(), verts.Normal[v].Y(), verts.Normal[v].Z())
		}
		if verts.Color != nil && v < len(verts.Color) {
			data = append(data,
				verts.Color[v].X(), verts.Color[v].Y(), verts.Color[v].Z(), verts.Color[v].W())
		}
	}
	return data, nil
}

func (verts *Vertices) BindAttributes() {
	stride := verts.Stride()

	bind := func(attr uint32, elems int, glType uint32, offset int) {
		gl.EnableVertexAttribArray(attr)
		gl.VertexAttribPointerWithOffset(attr, int32(elems), glType, false, int32(stride), uintptr(offset))
	}

	ofs := 0
	if verts.Pos != nil && len(verts.Pos) > 0 {
		bind(ATTR_POS, 3, gl.FLOAT, ofs)
		ofs += SIZEOF_POS
	}
	if verts.TexCoord != nil && len(verts.TexCoord) > 0 {
		bind(ATTR_TEXCOORD, 2, gl.FLOAT, ofs)
		ofs += SIZEOF_TEXCOORD
	}
	if verts.Normal != nil && len(verts.Normal) > 0 {
		bind(ATTR_NORMAL, 3, gl.FLOAT, ofs)
		ofs += SIZEOF_NORMAL
	}
	if verts.Color != nil && len(verts.Color) > 0 {
		bind(ATTR_COLOR, 4, gl.FLOAT, ofs)
		ofs += SIZEOF_COLOR
	}
}

type Group struct {
	Offset, Length int
}

type Mesh struct {
	verts Vertices
	inds  []uint32

	tris   []math2.Triangle // Mathematical representation of the triangles (will be empty until Triangles() is called)
	bbox   math2.Box        // Bounding box over vertices. Lazily evaluated.
	groups map[string]Group

	uploaded              bool
	vertBuffer, idxBuffer uint32 //VBOs
	vertArray             uint32 //VAO
}

func CreateMesh(verts Vertices, inds []uint32) *Mesh {
	mesh := &Mesh{
		verts:  verts,
		inds:   inds,
		tris:   nil,
		groups: make(map[string]Group, 0),
	}
	return mesh
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

// Returns the mathematical triangles that make up the mesh (lazily evaluated).
func (m *Mesh) Triangles() []math2.Triangle {
	if m.tris == nil {
		// Determine the triangles from the indices & vertex positions.
		m.tris = make([]math2.Triangle, len(m.inds)/3)
		for t := range m.tris {
			m.tris[t] = math2.Triangle{
				m.verts.Pos[m.inds[t*3+0]],
				m.verts.Pos[m.inds[t*3+1]],
				m.verts.Pos[m.inds[t*3+2]],
			}
		}
	}

	return m.tris
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

func (m *Mesh) Bind() {
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
	gl.DrawElementsWithOffset(gl.TRIANGLES, int32(len(m.inds)), gl.UNSIGNED_INT, 0)
}

func (m *Mesh) DrawGroup(name string) error {
	group, ok := m.groups[name]
	if !ok {
		return fmt.Errorf("Group not found")
	}
	gl.DrawElementsWithOffset(gl.TRIANGLES, int32(group.Length), gl.UNSIGNED_INT, uintptr(group.Offset)*unsafe.Sizeof(m.inds[0]))
	return nil
}

func (m *Mesh) Free() {
	gl.DeleteBuffers(1, &m.vertBuffer)
	gl.DeleteBuffers(1, &m.idxBuffer)
	gl.DeleteVertexArrays(1, &m.vertArray)
}
