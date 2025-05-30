package geom

import (
	"bufio"
	_ "fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type OBJIndex [3]int // The three indices into the .obj's vertex elements (1-based)

type OBJFace [4]OBJIndex // May be a quad or a triangle. For triangles, last index is all -1.

type OBJMaterial struct {
	faces []OBJFace
}

type OBJ struct {
	pos       [][3]float32
	tex       [][2]float32
	norm      [][3]float32
	materials map[string]*OBJMaterial
}

func LoadOBJMesh(path string) (*Mesh, error) {
	file, err := assets.GetFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	verts := Vertices{
		Pos:      make([]mgl32.Vec3, 0),
		TexCoord: make([]mgl32.Vec2, 0),
		Normal:   make([]mgl32.Vec3, 0),
		Color:    make([]mgl32.Vec4, 0),
	}
	inds := make([]uint32, 0)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	obj := OBJ{
		pos:       make([][3]float32, 0, 16),
		tex:       make([][2]float32, 0, 16),
		norm:      make([][3]float32, 0, 16),
		materials: make(map[string]*OBJMaterial),
	}

	// Takes a .obj face vertex definition "A/B/C"
	// And stores the index into `verts` of the corresponding mesh vertex.
	vertSet := make(map[OBJIndex]int)

	materialName := ""
	obj.materials[materialName] = &OBJMaterial{
		faces: make([]OBJFace, 0),
	}

	// Scan .obj file
	for scanner.Scan() {
		line := scanner.Text()
		tokens := strings.Split(line, " ")
		switch tokens[0] {
		case "v":
			var x, y, z float64
			x, err = strconv.ParseFloat(tokens[1], 32)
			if err != nil {
				return nil, err
			}
			y, err = strconv.ParseFloat(tokens[2], 32)
			if err != nil {
				return nil, err
			}
			z, err = strconv.ParseFloat(tokens[3], 32)
			if err != nil {
				return nil, err
			}
			obj.pos = append(obj.pos, [3]float32{float32(x), float32(y), float32(z)})
		case "vt":
			var u, v float64
			u, err = strconv.ParseFloat(tokens[1], 32)
			if err != nil {
				return nil, err
			}
			v, err = strconv.ParseFloat(tokens[2], 32)
			if err != nil {
				return nil, err
			}
			obj.tex = append(obj.tex, [2]float32{float32(u), float32(v)})
		case "vn":
			var x, y, z float64
			x, err = strconv.ParseFloat(tokens[1], 32)
			if err != nil {
				return nil, err
			}
			y, err = strconv.ParseFloat(tokens[2], 32)
			if err != nil {
				return nil, err
			}
			z, err = strconv.ParseFloat(tokens[3], 32)
			if err != nil {
				return nil, err
			}
			obj.norm = append(obj.norm, [3]float32{float32(x), float32(y), float32(z)})
		case "usemtl":
			materialName = tokens[1]
			_, ok := obj.materials[materialName]
			if !ok {
				obj.materials[materialName] = &OBJMaterial{
					faces: make([]OBJFace, 0),
				}
			}
		case "f":
			if len(tokens) > 5 || len(tokens) < 4 {
				log.Println("Error: OBJ loader found unsupported polygon type.")
				continue
			}
			var face OBJFace
			face[3] = OBJIndex{-1, -1, -1} // Marks a triangle unless overwritten
			for i := 1; i < len(tokens); i++ {
				// Parse index
				var faceTokens = strings.Split(tokens[i], "/")
				var idx OBJIndex
				for t, str := range faceTokens {
					if len(str) > 0 {
						idx[t], err = strconv.Atoi(str)
					} else {
						idx[t] = -1
					}
				}
				face[i-1] = idx

				// Create vertices for each face's vertex
				_, ok := vertSet[idx]
				if !ok {
					v := len(verts.Pos)
					vertSet[idx] = v

					if idx[0] > 0 { // Position
						verts.Pos = append(verts.Pos,
							mgl32.Vec3(obj.pos[idx[0]-1]))
					}
					if idx[1] > 0 { // Tex coord
						verts.TexCoord = append(verts.TexCoord,
							mgl32.Vec2(obj.tex[idx[1]-1]))
					}
					if idx[2] > 0 { // Normal
						verts.Normal = append(verts.Normal,
							mgl32.Vec3(obj.norm[idx[2]-1]))
					}

					// Assign white as default vertex color.
					verts.Color = append(verts.Color, mgl32.Vec4{1.0, 1.0, 1.0, 1.0})
				}
			}
			obj.materials[materialName].faces = append(obj.materials[materialName].faces, face)
		}
	}

	meshGroups := make(map[string]Group)
	for name, group := range obj.materials {
		meshGroup := Group{Offset: len(inds)}
		for _, face := range group.faces {
			// Add indices from faces
			if face[3][0] == -1 {
				// Triangle
				inds = append(inds,
					uint32(vertSet[face[0]]),
					uint32(vertSet[face[1]]),
					uint32(vertSet[face[2]]),
				)
				meshGroup.Length += 3
			} else {
				// Quad (two triangles)
				inds = append(inds,
					uint32(vertSet[face[0]]),
					uint32(vertSet[face[1]]),
					uint32(vertSet[face[2]]),
					uint32(vertSet[face[2]]),
					uint32(vertSet[face[3]]),
					uint32(vertSet[face[0]]),
				)
				meshGroup.Length += 6
			}
		}
		meshGroups[name] = meshGroup
	}
	mesh := CreateMesh(verts, inds)
	for name, group := range meshGroups {
		mesh.SetGroup(name, group)
	}
	mesh.Upload()

	log.Printf("Loaded OBJ file at %v.\n", path)
	return mesh, err
}
