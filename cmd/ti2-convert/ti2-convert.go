package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
)

func removeTextureTags(textureName string) string {
	suffixSlice := textureName[:]
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
	return strings.TrimSuffix(textureName, suffixSlice)
}

// Converts Total Invasion II texture name to Total Invasion 22 file path.
func translateTextureName(category, ti2Name string) string {
	var folderName string
	if strings.Contains(ti2Name, "space") {
		folderName = "space/"
	} else if strings.Contains(ti2Name, "prismir") {
		folderName = "prismir/"
	}

	ti2Name = removeTextureTags(ti2Name)

	// Direct texture name replacements
	if strings.HasPrefix(ti2Name, "charredgrass") {
		ti2Name = "charred_grass"
	} else {
		replacements := map[string]string{
			"grass_arrow0":    "grass",
			"grass_arrow1":    "grass",
			"balloonstand":    "balloon_stand",
			"cartonofeggs":    "egg_carton",
			"dopefish":        "dopefish_statue",
			"chickencannon":   "chicken_cannon",
			"grenadelauncher": "grenade_launcher",
			"plasmavials":     "plasma_vials",
			"exitsign":        "exit_sign",
		}
		if newName, ok := replacements[ti2Name]; ok {
			ti2Name = newName
		}
	}

	return fmt.Sprintf("assets/textures/%v/%v%v.png", category, folderName, ti2Name)
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

	tilesToAdd := make([]tileToAdd, 0, 1024)
	entsToAdd := make([]te3.Ent, 0, 256)

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

		textureIndex := addTexture(translateTextureName("tiles", tokens[3]))

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
			// Add dynamic tile entities
			ent := te3.Ent{
				Display: te3.ENT_DISPLAY_MODEL,
				Radius:  1.0,
				Model:   modelList[modelIndex],
				Texture: textureList[textureIndex],
				Angles:  [3]float32{0.0, float32(yaw * 90), 0.0},
				Color:   [3]uint8{255, 255, 255},
				Position: [3]float32{
					float32(x)*te3.GRID_SPACING + te3.HALF_GRID_SPACING,
					te3.GRID_SPACING + te3.HALF_GRID_SPACING,
					float32(z)*te3.GRID_SPACING + te3.HALF_GRID_SPACING,
				},
				Properties: map[string]string{},
			}

			switch flag {
			case 1, 2, 5, 8, 9, 10: // Moving door like objects
				ent.Properties["type"] = "door"
				if link > 0 {
					ent.Properties["link"] = strconv.FormatInt(link, 10)
				}
				switch flag {
				case 1, 8, 2, 9: // Doors
					if strings.Contains(tokens[3], "spacedoor") {
						// Space doors move up instead of to the side
						ent.Properties["direction"] = "up"
					} else {
						ent.Properties["direction"] = "right"
					}

					if flag == 8 || flag == 9 {
						// Assign keys to locked doors
						switch link {
						case 0:
							ent.Properties["key"] = "blue"
						case 1:
							ent.Properties["key"] = "brown"
						case 2:
							ent.Properties["key"] = "yellow"
						case 3:
							ent.Properties["key"] = "gray"
						}
						delete(ent.Properties, "link")
					}
				case 5: // Push walls
					ent.Properties["direction"] = "backward"
					ent.Properties["distance"] = "4.0"
					ent.Properties["wait"] = "inf"
					ent.Properties["activateSound"] = "secretwall.wav"
					delete(ent.Properties, "link")
				case 10: // Disappearing walls
					ent.Properties["direction"] = "down"
					ent.Properties["distance"] = "4.0"
					ent.Properties["activateSound"] = ""
					ent.Properties["blockUse"] = "true"
					ent.Properties["wait"] = "inf"
				}
			case 3: // Switch
				ent.Properties["type"] = "switch"
				ent.Properties["link"] = strconv.FormatInt(link, 10)
			case 4, 11, 12, 13: // Invisible trigger volumes
				ent.Properties["type"] = "trigger"
				ent.Properties["link"] = strconv.FormatInt(link, 10)
				switch flag {
				case 4: // Teleporter
					ent.Properties["action"] = "teleport"
				case 11: // Trigger for doors / secrets
					if link == 255 {
						ent.Properties["action"] = "secret"
						delete(ent.Properties, "link")
					} else {
						ent.Properties["action"] = "activate"
					}
				case 12: // Level exit
					ent.Properties["action"] = "end level"
					ent.Properties["level"] = "TODO!"
					delete(ent.Properties, "link")
				case 13: // Conveyor belt
					ent.Properties["action"] = "push"
					var dir string
					switch link {
					case 0:
						dir = "forward"
					case 1:
						dir = "right"
					case 2:
						dir = "backward"
					case 3:
						dir = "left"
					}
					ent.Properties["direction"] = dir
				}
			}

			entsToAdd = append(entsToAdd, ent)
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

		textureIndex := addTexture(translateTextureName("tiles", tokens[5]))

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

	tiScanner.Scan() // THINGS
	tiScanner.Scan() // Ent count
	entCount, err := strconv.ParseInt(tiScanner.Text(), 10, 64)
	if err != nil {
		panic(err)
	}

	// Calculate grid bounds
	gridWidth := int(maxX - minX + 1)
	gridLength := int(maxZ - minZ + 1)
	gridLayerSize := gridWidth * gridLength
	grid := slices.Repeat([]te3.Tile{{ShapeID: -1}}, gridWidth*3*gridLength)
	for _, pair := range tilesToAdd {
		grid[(pair.coords[0]-int(minX))+((pair.coords[2]-int(minZ))*gridWidth)+(pair.coords[1]*gridLayerSize)] = pair.Tile
	}

	// Correct positions for tile entities
	for i := range entsToAdd {
		entsToAdd[i].Position[0] -= float32(minX) * te3.GRID_SPACING
		entsToAdd[i].Position[2] -= float32(minZ) * te3.GRID_SPACING
	}

	for range entCount {
		tiScanner.Scan()
		tokens := strings.Split(tiScanner.Text(), ",")

		x, err := strconv.ParseInt(tokens[0], 10, 64)
		if err != nil {
			panic(err)
		}
		x = (x / 16) - minX

		z, err := strconv.ParseInt(tokens[1], 10, 64)
		if err != nil {
			panic(err)
		}
		z = (z / 16) - minZ

		if x < 0 || z < 0 || x >= int64(gridWidth) || z >= int64(gridLength) {
			fmt.Printf("Ent found out of bounds: [%v, %v]\n", x, z)
			continue
		}

		angleIndex, err := strconv.ParseInt(tokens[4], 10, 64)
		if err != nil {
			panic(err)
		}

		ent := te3.Ent{
			Radius: 0.7,
			Angles: [3]float32{
				0.0, 270.0 - float32(angleIndex*45), 0.0,
			},
			Color: [3]uint8{255, 255, 255},
			Position: [3]float32{
				float32(x)*te3.GRID_SPACING + te3.HALF_GRID_SPACING,
				te3.GRID_SPACING + te3.HALF_GRID_SPACING,
				float32(z)*te3.GRID_SPACING + te3.HALF_GRID_SPACING,
			},
			Display:    te3.ENT_DISPLAY_SPHERE,
			Properties: map[string]string{},
		}

		entType, err := strconv.ParseInt(tokens[2], 10, 64)
		if err != nil {
			panic(err)
		}

		shouldBeSprite := false
		strippedTexName := removeTextureTags(tokens[3])

		switch entType {
		case 0: // Player
			ent.Display = te3.ENT_DISPLAY_SPRITE
			ent.Texture = "assets/textures/sprites/segan.png"
			ent.Properties["type"] = "player"
		case 1: // Prop
			ent.Properties["type"] = "prop"
			ent.Properties["prop"] = strippedTexName
			shouldBeSprite = true
		case 2, 3: // Item or weapon
			ent.Properties["type"] = "item"
			ent.Properties["item"] = strippedTexName
			shouldBeSprite = true
		case 4: // Wraith
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "wraith"
			ent.Display = te3.ENT_DISPLAY_SPRITE
			ent.Texture = "assets/textures/sprites/wraith.png"
		case 5: // Fire wraith
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "fire wraith"
			ent.Display = te3.ENT_DISPLAY_SPRITE
			ent.Texture = "assets/textures/sprites/fire_wraith.png"
		case 6, 14: // Dummkopf
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "dummkopf"
			ent.Properties["name"] = "dummkopf"
		case 7: // Mother wraith / Caco wraith
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "mother wraith"
			ent.Display = te3.ENT_DISPLAY_SPRITE
			ent.Texture = "assets/textures/sprites/mother_wraith.png"
		case 8: // Prisrak
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "prisrak"
			ent.Properties["name"] = "prisrak"
		case 9: // Providence
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "providence"
			ent.Properties["name"] = "providence"
		case 10: // Fundie
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "fundie"
			ent.Properties["name"] = "fundie"
		case 11: // Banshee
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "banshee"
			ent.Properties["name"] = "banshee"
		case 12: // Mutant wraith
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "mutant wraith"
			ent.Properties["name"] = "mutant wraith"
		case 13: // Mecha
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "mecha wraith"
			ent.Properties["name"] = "mecha wraith"
		case 15: // Tophat demon
			ent.Properties["type"] = "enemy"
			ent.Properties["enemy"] = "tophat demon"
			ent.Properties["name"] = "tophat demon"
		default:
			// Discard unused entity types
			continue
		}

		if shouldBeSprite {
			texPath := translateTextureName("sprites", strippedTexName)
			if imgFile, err := os.Open(texPath); err == nil {
				defer imgFile.Close()

				ent.Display = te3.ENT_DISPLAY_SPRITE

				// Scale according to image dimensions
				img, _, err := image.Decode(imgFile)
				if err != nil {
					panic(err)
				}
				ent.Radius = float32(img.Bounds().Dy()) / 64.0

				ent.Texture = texPath
			} else {
				ent.Properties["name"] = strippedTexName
			}
		}

		entsToAdd = append(entsToAdd, ent)
	}

	te3Map := te3.TE3File{
		Ents: entsToAdd,
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
