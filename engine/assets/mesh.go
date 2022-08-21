package assets

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	
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
		return nil, fmt.Errorf("Vertices has no position data!")
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
	Verts                 Vertices
	Inds                  []uint32

	groups                map[string]Group

	uploaded              bool
	vertBuffer, idxBuffer uint32 //VBOs
	vertArray             uint32 //VAO
}

func CreateMesh(verts Vertices, inds []uint32) *Mesh {
	mesh := &Mesh{
		Verts:      verts,
		Inds:       inds,
		groups:     make(map[string]Group, 0),
	}
	mesh.Upload()
	return mesh
}

func (m *Mesh) SetGroup(name string, group Group) {
	m.groups[name] = group
}

func (m *Mesh) HasGroup(name string) bool {
	_, ok := m.groups[name]
	return ok
}

func (m *Mesh) GetGroupNames() []string {
	out := make([]string, 0, len(m.groups))
	for name := range m.groups {
		out = append(out, name)
	}
	return out
}

func (m *Mesh) Bind() {
	if !m.uploaded {
		m.Upload()
	}

	gl.BindVertexArray(m.vertArray)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vertBuffer)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, m.idxBuffer)
	m.Verts.BindAttributes()
}

func (m *Mesh) Upload() {
	if m.uploaded {
		m.Free()
	}

	//Flatten the vertex array into a series of floats
	data, err := m.Verts.Flatten()
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
		len(m.Inds)*int(unsafe.Sizeof(m.Inds[0])), //Size in bytes of buffer data
		gl.Ptr(m.Inds),
		gl.STATIC_DRAW)

	m.uploaded = true
}

func (m *Mesh) DrawAll() {
	gl.DrawElements(gl.TRIANGLES, int32(len(m.Inds)), gl.UNSIGNED_INT, gl.PtrOffset(0))
}

func (m *Mesh) DrawGroup(name string) error {
	group, ok := m.groups[name]
	if !ok {
		return fmt.Errorf("Group not found")
	}
	gl.DrawElementsWithOffset(gl.TRIANGLES, int32(group.Length), gl.UNSIGNED_INT, uintptr(group.Offset) * unsafe.Sizeof(m.Inds[0]))
	return nil
}

func (m *Mesh) Free() {
	gl.DeleteBuffers(1, &m.vertBuffer)
	gl.DeleteBuffers(1, &m.idxBuffer)
	gl.DeleteVertexArrays(1, &m.vertArray)
}
