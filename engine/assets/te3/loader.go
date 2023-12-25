package te3

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

const GRID_SPACING = 2.0
const HALF_GRID_SPACING = GRID_SPACING / 2.0

type TE3File struct {
	Ents     []Ent
	Tiles    Tiles
	filePath string
}

type EntDisplay uint8

const (
	ENT_DISPLAY_SPHERE EntDisplay = iota
	ENT_DISPLAY_MODEL
	ENT_DISPLAY_SPRITE
)

type Ent struct {
	Angles         [3]float32
	Color          [3]uint8
	Position       [3]float32
	Radius         float32
	Texture, Model string
	Display        EntDisplay
	Properties     map[string]string
}

type Tiles struct {
	Data                  []Tile
	Width, Height, Length int
	Textures              []string
	Shapes                []string
}

type ShapeID int32
type TextureID int32

type Tile struct {
	ShapeID   ShapeID
	Yaw       int32 //Yaw in whole number of degrees
	TextureID TextureID
	Pitch     int32 //Pitch in whole number of degrees
}

// Returns the rotation matrix based off of the tile's yaw and pitch values.
func (t *Tile) GetRotationMatrix() mgl32.Mat4 {
	return mgl32.HomogRotate3DY(mgl32.DegToRad(float32(-t.Yaw))).Mul4(
		mgl32.HomogRotate3DX(mgl32.DegToRad(float32(-t.Pitch))))
}

// Parses tile data from JSON file (manually, since Tile has non-decoded fields)
func (tiles *Tiles) UnmarshalJSON(b []byte) error {
	//Get JSON data
	var jData map[string]any
	if err := json.Unmarshal(b, &jData); err != nil {
		return err
	}

	tiles.Width = int(jData["width"].(float64))
	tiles.Height = int(jData["height"].(float64))
	tiles.Length = int(jData["length"].(float64))

	//Parse and convert texture paths
	textures := jData["textures"].([]any)
	tiles.Textures = make([]string, len(textures))
	for t, tex := range textures {
		tiles.Textures[t] = tex.(string)
	}

	//Parse and convert model paths
	shapes := jData["shapes"].([]any)
	tiles.Shapes = make([]string, len(shapes))
	for s, shape := range shapes {
		tiles.Shapes[s] = shape.(string)
	}

	//Convert tile data from base64
	tileString := jData["data"].(string)
	tileBytes, err := base64.StdEncoding.DecodeString(tileString)
	if err != nil {
		return err
	}

	//Parse bytes as tile array
	reader := bytes.NewReader(tileBytes)
	tiles.Data = make([]Tile, 0)
	var tile Tile
	for t := int64(0); t < reader.Size(); t++ {
		if binary.Read(reader, binary.LittleEndian, &tile.ShapeID) != io.ErrUnexpectedEOF &&
			binary.Read(reader, binary.LittleEndian, &tile.Yaw) != io.ErrUnexpectedEOF &&
			binary.Read(reader, binary.LittleEndian, &tile.TextureID) != io.ErrUnexpectedEOF &&
			binary.Read(reader, binary.LittleEndian, &tile.Pitch) != io.ErrUnexpectedEOF {

			if tile.ShapeID < 0 {
				//Negative shape ID represents a run of empty tiles
				for r := 0; r < -int(tile.ShapeID); r++ {
					tiles.Data = append(tiles.Data, Tile{ShapeID: ShapeID(-1), Yaw: 0, TextureID: TextureID(-1), Pitch: 0})
				}
			} else {
				tiles.Data = append(tiles.Data, tile)
			}
		} else {
			return io.ErrUnexpectedEOF
		}
	}
	return nil
}

// Loads a Total Editor 3 map file into a data structure
func LoadTE3File(assetPath string) (*TE3File, error) {
	te3, err := assets.LoadAndUnmarshalJSON[TE3File](assetPath)
	te3.filePath = assetPath
	if err == nil {
		log.Println("Loaded TE3 file", assetPath)
	}
	return te3, err
}

func (te3 *TE3File) FilePath() string {
	return te3.filePath
}

// Returns the integer grid position (x, y, z) from the given flat index into the Data array
func (tiles *Tiles) UnflattenGridPos(index int) (int, int, int) {
	return (index % tiles.Width), (index / (tiles.Width * tiles.Length)), ((index / tiles.Width) % tiles.Length)
}

// Returns the flat index into the Data array for the given integer grid position (does not validate).
func (tiles *Tiles) FlattenGridPos(x, y, z int) int {
	return x + (z * tiles.Width) + (y * tiles.Width * tiles.Length)
}

func (tiles *Tiles) WorldToGridPos(worldPos mgl32.Vec3) (int, int, int) {
	var out [3]int
	for i := range out {
		out[i] = int(worldPos[i] / GRID_SPACING)
	}
	return out[0], out[1], out[2]
}

func (tiles *Tiles) OutOfBounds(x, y, z int) bool {
	return x < 0 || y < 0 || z < 0 || x >= tiles.Width || y >= tiles.Height || z >= tiles.Length
}

func (tiles *Tiles) GridSpacing() float32 {
	return GRID_SPACING
}

func (tiles *Tiles) GridToWorldPos(i, j, k int, center bool) mgl32.Vec3 {
	out := mgl32.Vec3{
		float32(i) * GRID_SPACING,
		float32(j) * GRID_SPACING,
		float32(k) * GRID_SPACING,
	}
	if center {
		out[0] += GRID_SPACING / 2.0
		out[1] += GRID_SPACING / 2.0
		out[2] += GRID_SPACING / 2.0
	}
	return out
}

// Returns the first entity in the map with the given key value pair in its properties.
func (te3 *TE3File) FindEntWithProperty(key, value string) (Ent, error) {
	for _, ent := range te3.Ents {
		if val, ok := ent.Properties[key]; ok && val == value {
			return ent, nil
		}
	}
	return Ent{}, fmt.Errorf("could not find entity with property tuple (%s, %s)", key, value)
}

// Returns an array of entities in the map with the given key value pair in their properties.
func (te3 *TE3File) FindEntsWithProperty(key, value string) []Ent {
	out := make([]Ent, 0)
	for _, ent := range te3.Ents {
		if val, ok := ent.Properties[key]; ok && val == value {
			out = append(out, ent)
		}
	}
	return out
}

// Extracts and parses the value of a float property.
func (ent *Ent) FloatProperty(key string) (float32, error) {
	prop, ok := ent.Properties[key]
	if !ok {
		return 0.0, fmt.Errorf("ent property not found: %v", key)
	}
	valF64, err := strconv.ParseFloat(prop, 32)
	if err != nil {
		return 0.0, err
	}
	return float32(valF64), nil
}

func (ent *Ent) BoolProperty(key string) (bool, error) {
	prop, ok := ent.Properties[key]
	if !ok {
		return false, fmt.Errorf("ent property not found: %v", prop)
	}
	return strconv.ParseBool(prop)
}
