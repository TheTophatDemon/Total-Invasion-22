package geom

import (
	"fmt"

	"github.com/go-gl/gl/v2.1/gl"
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
