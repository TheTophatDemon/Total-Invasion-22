package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
)

// Converts Total Invasion II texture name to Total Invasion 22 file path.
func translateTextureName(ti2Name string) string {
	var folderName string
	if strings.Contains(ti2Name, "space") {
		folderName = "space/"
	} else if strings.Contains(ti2Name, "prismir") {
		folderName = "prismir/"
	}

	// Cut out metadata suffixes
	suffixSlice := ti2Name[:]
findSuffix:
	for len(suffixSlice) > 1 {
		underSc := strings.IndexRune(suffixSlice[1:], '_')
		if underSc < 0 {
			underSc = len(suffixSlice)
		} else {
			underSc++
		}
		switch suffixSlice[:underSc] {
		case "_animated", "_invisible", "_notsolid":
			break findSuffix
		}
		suffixSlice = suffixSlice[underSc:]
	}
	ti2Name = strings.TrimSuffix(ti2Name, suffixSlice)

	//TODO: direct texture name replacements

	return fmt.Sprintf("assets/textures/tiles/%v%v.png", folderName, ti2Name)
}

func main() {
	if len(os.Args) != 3 {
		panic("Need two args: input path (.ti) and output path (.te3)")
	}

	inputPath := os.Args[1]
	tiFile, err := os.Open(inputPath)
	if err != nil {
		panic(err)
	}
	defer tiFile.Close()

	type tileToAdd struct {
		te3.Tile
		coords [3]int
	}

	const (
		IDX_CUBE = iota
		IDX_PANEL
		IDX_BARS
		IDX_DOOR
		IDX_MARKER
	)

	modelList := []string{
		IDX_CUBE:   "assets/models/shapes/cube.obj",
		IDX_PANEL:  "assets/models/shapes/panel.obj",
		IDX_BARS:   "assets/models/shapes/bars.obj",
		IDX_DOOR:   "assets/models/shapes/door.obj",
		IDX_MARKER: "assets/models/shapes/cube_marker.obj",
	}
	textureList := make([]string, 0, 64)

	addTexture := func(texturePath string) int {
		textureIndex := slices.Index(textureList, texturePath)
		if textureIndex < 0 {
			textureIndex = len(textureList)
			textureList = append(textureList, texturePath)
		}
		return textureIndex
	}

	tilesToAdd := make([]tileToAdd, 0, 256)

	tiScanner := bufio.NewScanner(tiFile)
	tiScanner.Scan() // TILES
	tiScanner.Scan() // Tile count
	wallCount, err := strconv.ParseInt(tiScanner.Text(), 10, 64)
	if err != nil {
		panic(err)
	}

	var minX int64 = math.MaxInt64
	var minZ int64 = math.MaxInt64
	var maxX int64 = math.MinInt64
	var maxZ int64 = math.MinInt64

	for range wallCount {
		tiScanner.Scan()
		tokens := strings.Split(tiScanner.Text(), ",")

		x, err := strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			panic(err)
		}
		x /= 16
		minX = min(x, minX)
		maxX = max(x, maxX)

		z, err := strconv.ParseInt(tokens[1], 10, 64)
		if err != nil {
			panic(err)
		}
		z /= 16
		minZ = min(z, minZ)
		maxZ = max(z, maxZ)

		textureIndex := addTexture(translateTextureName(tokens[3]))

		flag, err := strconv.ParseInt(tokens[4], 10, 64)
		if err != nil {
			panic(err)
		}
		link, err := strconv.ParseInt(tokens[5], 10, 64)
		if err != nil {
			panic(err)
		}

		var yaw uint8
		switch flag {
		case 7: // Panel
			yaw = uint8(link % 2)
		case 2, 9: // Z-axis doors
			yaw = 1
		case 5: // Secret walls
			yaw = uint8((1 + link) % 4)
		}

		// Set shape depending on tile type
		modelIndex := IDX_CUBE
		switch flag {
		case 1, 2, 7, 8, 9:
			modelIndex = IDX_PANEL
		}
		switch tokens[3] {
		case "bars_invisible", "brokenbars_invisible_notsolid":
			modelIndex = IDX_BARS
		case "door", "spacedoor", "spacedoor_dark", "prismirdoor":
			modelIndex = IDX_DOOR
		case "invisible":
			modelIndex = IDX_MARKER
		}

		if flag > 0 && flag != 6 && flag != 7 {
			//TODO: Add tile entities
		} else {
			tilesToAdd = append(tilesToAdd, tileToAdd{
				Tile: te3.Tile{
					ShapeID:    te3.ShapeID(modelIndex),
					TextureIDs: [2]te3.TextureID{te3.TextureID(textureIndex), te3.TextureID(textureIndex)},
					Yaw:        yaw,
					Pitch:      0,
				},
				coords: [3]int{int(x), 1, int(z)},
			})
		}
	}

	tiScanner.Scan() // SECTORS
	tiScanner.Scan() // Floor count
	floorCount, err := strconv.ParseInt(tiScanner.Text(), 10, 64)
	if err != nil {
		panic(err)
	}

	for range floorCount {
		tiScanner.Scan()
		tokens := strings.Split(tiScanner.Text(), ",")

		x, err := strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			panic(err)
		}
		x /= 16
		minX = min(x, minX)
		maxX = max(x, maxX)

		z, err := strconv.ParseInt(tokens[1], 10, 64)
		if err != nil {
			panic(err)
		}
		z /= 16
		minZ = min(z, minZ)
		maxZ = max(z, maxZ)

		textureIndex := addTexture(translateTextureName(tokens[5]))

		ceilingFlag, err := strconv.ParseInt(tokens[4], 10, 64)
		if err != nil {
			panic(err)
		}

		tilesToAdd = append(tilesToAdd, tileToAdd{
			coords: [3]int{int(x), int(ceilingFlag * 2), int(z)},
			Tile: te3.Tile{
				ShapeID:    IDX_CUBE,
				TextureIDs: [2]te3.TextureID{te3.TextureID(textureIndex), te3.TextureID(textureIndex)},
			},
		})
	}

	// Calculate grid bounds
	gridWidth := int(maxX - minX + 1)
	gridLength := int(maxZ - minZ + 1)
	gridLayerSize := gridWidth * gridLength
	grid := slices.Repeat([]te3.Tile{{ShapeID: -1}}, gridWidth*3*gridLength)
	for _, pair := range tilesToAdd {
		grid[(pair.coords[0]-int(minX))+((pair.coords[2]-int(minZ))*gridWidth)+(pair.coords[1]*gridLayerSize)] = pair.Tile
	}

	te3Map := te3.TE3File{
		Ents: []te3.Ent{},
		Tiles: te3.Tiles{
			Textures: textureList,
			Shapes:   modelList,
			Width:    gridWidth,
			Height:   3,
			Length:   gridLength,
			Data:     grid,
		},
	}
	te3Map.Meta.Editor = "Total Editor"
	te3Map.Meta.Version = "3.2"

	jsonData, err := json.Marshal(&te3Map)
	if err != nil {
		panic(err)
	}
	te3File, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	defer te3File.Close()
	_, err = te3File.Write(jsonData)
	if err != nil {
		panic(err)
	}
}
