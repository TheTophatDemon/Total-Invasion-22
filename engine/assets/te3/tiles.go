package te3

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

const GRID_SPACING = 2.0
const HALF_GRID_SPACING = GRID_SPACING / 2.0

type Tiles struct {
	Data                  []Tile
	Width, Height, Length int
	Textures              []string
	Shapes                []string
}

type ShapeID int16
type TextureID int16

type Tile struct {
	ShapeID    ShapeID
	TextureIDs [2]TextureID
	Yaw        uint8 // 0-3 for each 90 degree increment.
	Pitch      uint8 // 0-3 for each 90 degree increment.
}

// Returns the rotation matrix based off of the tile's yaw and pitch values.
func (t *Tile) GetRotationMatrix() mgl32.Mat4 {
	return mgl32.HomogRotate3DY(float32(-t.Yaw) * math2.HALF_PI).Mul4(
		mgl32.HomogRotate3DX(float32(-t.Pitch) * math2.HALF_PI))
}

// Parses tile data from JSON file (manually, since Tile has non-decoded fields)
func (tiles *Tiles) UnmarshalJSON(b []byte) error {
	// Get JSON data
	var jData map[string]any
	if err := json.Unmarshal(b, &jData); err != nil {
		return err
	}

	tiles.Width = int(jData["width"].(float64))
	tiles.Height = int(jData["height"].(float64))
	tiles.Length = int(jData["length"].(float64))

	// Parse and convert texture paths
	textures := jData["textures"].([]any)
	tiles.Textures = make([]string, len(textures))
	for t, tex := range textures {
		tiles.Textures[t] = tex.(string)
	}

	// Parse and convert model paths
	shapes := jData["shapes"].([]any)
	tiles.Shapes = make([]string, len(shapes))
	for s, shape := range shapes {
		tiles.Shapes[s] = shape.(string)
	}

	// Convert tile data from base64
	tileString := jData["data"].(string)
	tileBytes, err := base64.StdEncoding.DecodeString(tileString)
	if err != nil {
		return err
	}

	// Parse bytes as tile array
	reader := bytes.NewReader(tileBytes)
	tiles.Data = make([]Tile, tiles.Width*tiles.Height*tiles.Length)
	tileIndex := 0
	for tileIndex < len(tiles.Data) {
		var tile Tile
		if binary.Read(reader, binary.LittleEndian, &tile.ShapeID) == io.ErrUnexpectedEOF {
			return io.ErrUnexpectedEOF
		}

		if tile.ShapeID < 0 {
			// Negative shape ID represents a run of empty tiles
			for range -tile.ShapeID {
				tiles.Data[tileIndex] = Tile{ShapeID: ShapeID(-1)}
				tileIndex++
			}
		} else if binary.Read(reader, binary.LittleEndian, tile.TextureIDs[:]) != io.ErrUnexpectedEOF &&
			binary.Read(reader, binary.LittleEndian, &tile.Yaw) != io.ErrUnexpectedEOF &&
			binary.Read(reader, binary.LittleEndian, &tile.Pitch) != io.ErrUnexpectedEOF {

			tiles.Data[tileIndex] = tile
			tileIndex++
		} else {
			return io.ErrUnexpectedEOF
		}
	}
	return nil
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

func (tiles *Tiles) BBoxOfTile(i, j, k int) math2.Box {
	corner := tiles.GridToWorldPos(i, j, k, false)
	return math2.Box{
		Min: corner,
		Max: corner.Add(mgl32.Vec3{tiles.GridSpacing(), tiles.GridSpacing(), tiles.GridSpacing()}),
	}
}

func (tiles *Tiles) EraseTile(tileID int) {
	tiles.Data[tileID] = Tile{ShapeID: -1}
}
