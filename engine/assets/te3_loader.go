package assets

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"github.com/go-gl/mathgl/mgl32"
)

const GRID_SPACING = 2.0

type TE3File struct {
	Ents  []Ent
	Tiles Tiles
}

type Ent struct {
	Angles     [3]float32
	Color      [3]uint8
	Position   [3]float32
	Radius     float32
	Properties map[string]string
}

type TileData []Tile

type Tiles struct {
	Data                  TileData
	Width, Height, Length uint32
	Textures              []string
	Shapes                []string
}

type ShapeID   int32
type TextureID int32

type Tile struct {
	ShapeID   ShapeID
    Yaw       int32 //Yaw in whole number of degrees
    TextureID TextureID
    Pitch     int32 //Pitch in whole number of degrees
}

//Returns the rotation matrix based off of the tile's yaw and pitch values.
func (t *Tile) GetRotationMatrix() mgl32.Mat4 {
	return mgl32.HomogRotate3DY(mgl32.DegToRad(float32(-t.Yaw))).Mul4(
		mgl32.HomogRotate3DX(mgl32.DegToRad(float32(-t.Pitch))))
}

func (tileData *TileData) UnmarshalJSON(b []byte) error {
	//Get the base64 string from the JSON
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	
	//Convert from base64
	tileBytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	//Parse bytes as tile array
	reader := bytes.NewReader(tileBytes)
	*tileData = make([]Tile, 0)
	var tile Tile
	for err == nil {
		err = binary.Read(reader, binary.LittleEndian, &tile)
		if err != io.ErrUnexpectedEOF { 
			*tileData = append(*tileData, tile)
		} else {
			return err
		}
	}
	return nil
}

func LoadTE3File(assetPath string) (*TE3File, error) {
	te3, err := LoadAndUnmarshalJSON[TE3File](assetPath)
	if err == nil {
		log.Println("Loaded TE3 file", assetPath)
	}
	return te3, err
}

func (tiles *Tiles) GetGridPos(index int) (int, int, int) {
	return (index % int(tiles.Width)), (index / int(tiles.Width * tiles.Length)), ((index / int(tiles.Width)) % int(tiles.Length))
}

func (te3 *TE3File) BuildMesh() (*Mesh, error) {
	var err error
	
	mapVerts := Vertices{
		Pos: make([]mgl32.Vec3, 0),
		TexCoord: make([]mgl32.Vec2, 0),
		Normal: make([]mgl32.Vec3, 0),
		Color: nil,
	}
	mapInds := make([]uint32, 0)

	shapeMeshes := make([]*Mesh, len(te3.Tiles.Shapes))
	for i, path := range te3.Tiles.Shapes {
		shapeMeshes[i], err = GetMesh(path)
		if err != nil {
			log.Println("WARNING: Shape mesh at", path, "not found!")
		}
	}

	//Groups tile data indices by their texture
	groupTiles := make(map[TextureID][]int)

	//Fill out tile groups
	for t, tile := range te3.Tiles.Data {
		//Only visible tiles are grouped
		if tile.ShapeID < 0 || tile.TextureID < 0 {
			continue
		}
		group, ok := groupTiles[tile.TextureID]
		if !ok {
			group = make([]int, 0, 16)
		}
		groupTiles[tile.TextureID] = append(group, t)
	}

	meshGroups := make([]Group, 0, len(groupTiles))
	meshGroupNames := make([]string, 0, len(groupTiles))

	//Add vertex data from tiles to map mesh
	for texID, tileIndices := range groupTiles {
		group := Group{ Offset: len(mapInds), Length: 0 }

		for _, t := range tileIndices {
			shapeMesh := shapeMeshes[te3.Tiles.Data[t].ShapeID]
			gridX, gridY, gridZ := te3.Tiles.GetGridPos(t)
			tileMatx := te3.Tiles.Data[t].GetRotationMatrix()
		
			tileVertexOffset := len(mapVerts.Pos)

			//Add the shape's vertex positions to the aggregate mesh, offset by the overall tile position
			for _, pos := range shapeMesh.Verts.Pos {
				pos = mgl32.TransformCoordinate(pos, tileMatx) //Rotate by tile orientation
				pos[0] += float32(gridX) * GRID_SPACING
				pos[1] += float32(gridY) * GRID_SPACING
				pos[2] += float32(gridZ) * GRID_SPACING
				mapVerts.Pos = append(mapVerts.Pos, pos)
			}

			//Append tex coordinates
			for _, uv := range shapeMesh.Verts.TexCoord {
				mapVerts.TexCoord = append(mapVerts.TexCoord, uv)
			}

			//Append normals, rotated by the tile orientation
			for _, rawNormal := range shapeMesh.Verts.Normal {
				normal := mgl32.TransformNormal(rawNormal, tileMatx)
				mapVerts.Normal = append(mapVerts.Normal, normal)
			}

			//Append indices, adding an offset to the tile's vertex data in the aggregate arrays.
			for _, ind := range shapeMesh.Inds {
				mapInds = append(mapInds, ind + uint32(tileVertexOffset))
			}

			group.Length += len(shapeMesh.Inds)
		}

		// group.Length = len(mapVerts.Pos) - group.Offset
		meshGroups = append(meshGroups, group)
		meshGroupNames = append(meshGroupNames, te3.Tiles.Textures[texID])
	}

	mesh := CreateMesh(mapVerts, mapInds)

	//Set group names to texture paths
	for g, group := range meshGroups {
		mesh.SetGroup(meshGroupNames[g], group)
	}

	return mesh, nil
}