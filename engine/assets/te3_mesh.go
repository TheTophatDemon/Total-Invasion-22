package assets

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

type CullGroup struct {
	complement int
	mask       uint32
	name       string
	dir        [3]int
}

var cullGroups = [...]CullGroup{
	{CULLGROUP_NONE, 0 << 0, "nocull", [3]int{0, 0, 0}},
	{CULLGROUP_DOWN, 1 << 0, "cullu", [3]int{0, 1, 0}},
	{CULLGROUP_UP, 1 << 1, "culld", [3]int{0, -1, 0}},
	{CULLGROUP_WEST, 1 << 2, "culle", [3]int{1, 0, 0}},
	{CULLGROUP_EAST, 1 << 3, "cullw", [3]int{-1, 0, 0}},
	{CULLGROUP_SOUTH, 1 << 4, "culln", [3]int{0, 0, -1}},
	{CULLGROUP_NORTH, 1 << 5, "culls", [3]int{0, 0, 1}},
}

// Indexes into the cullGroups array
const (
	CULLGROUP_NONE = iota
	CULLGROUP_UP
	CULLGROUP_DOWN
	CULLGROUP_EAST
	CULLGROUP_WEST
	CULLGROUP_NORTH
	CULLGROUP_SOUTH
)

var cullGroupFromName map[string]*CullGroup
var cullGroupFromDir map[[3]int]*CullGroup

func init() {
	cullGroupFromName = make(map[string]*CullGroup)
	cullGroupFromDir = make(map[[3]int]*CullGroup)

	for i, cg := range cullGroups {
		cullGroupFromName[cg.name] = &cullGroups[i]
		cullGroupFromDir[cg.dir] = &cullGroups[i]
	}

	cullGroupFromName[""] = &cullGroups[CULLGROUP_NONE]
}

func (tiles *Tiles) ShouldCull(tile Tile, gridX, gridY, gridZ int, testGroup *CullGroup) bool {
	//Don't cull if the tile doesn't have that cull group
	if (tile.cullGroup & testGroup.mask) == 0 {
		return false
	}

	//Neighboring tile's grid position
	nGridX := gridX + testGroup.dir[0]
	nGridY := gridY + testGroup.dir[1]
	nGridZ := gridZ + testGroup.dir[2]

	//If the neighbor is out of bounds then cull the tile
	if nGridX < 0 || nGridY < 0 || nGridZ < 0 || nGridX >= tiles.Width || nGridY >= tiles.Height || nGridZ >= tiles.Length {
		return false
	}

	neighbor := tiles.Data[tiles.FlattenGridPos(nGridX, nGridY, nGridZ)]

	//Don't cull when the neighbor is a different shape or if it is empty
	if neighbor.ShapeID != tile.ShapeID || neighbor.ShapeID < 0 || neighbor.TextureID < 0 {
		return false
	}

	//Cull if the neighbor has a complementary culling group
	return (neighbor.cullGroup & cullGroups[testGroup.complement].mask) != 0
}

func (te3 *TE3File) BuildMesh() (*Mesh, error) {
	var err error

	mapVerts := Vertices{
		Pos:      make([]mgl32.Vec3, 0),
		TexCoord: make([]mgl32.Vec2, 0),
		Normal:   make([]mgl32.Vec3, 0),
		Color:    nil,
	}
	mapInds := make([]uint32, 0)

	shapeMeshes := make([]*Mesh, len(te3.Tiles.Shapes))
	for i, path := range te3.Tiles.Shapes {
		shapeMeshes[i], err = GetMesh(path)
		if err != nil {
			return nil, fmt.Errorf("shape mesh at %s not found", path)
		}
	}

	//Groups tile data indices by their texture
	groupTiles := make(map[TextureID][]int)

	//Preprocess tiles
	for t, tile := range te3.Tiles.Data {
		//Only visible tiles are processed here
		if tile.ShapeID < 0 || tile.TextureID < 0 {
			continue
		}

		//Exclude tiles with invisible texture flag
		if GetTexture(te3.Tiles.Textures[tile.TextureID]).HasFlag("invisible") {
			continue
		}

		//Assign to group based on texture
		group, ok := groupTiles[tile.TextureID]
		if !ok {
			group = make([]int, 0, 16)
		}
		groupTiles[tile.TextureID] = append(group, t)

		//Calculate rotation matrix for later
		rotMatx := tile.GetRotationMatrix()
		te3.Tiles.Data[t].rotMatx = rotMatx

		//Assign culling groups
		te3.Tiles.Data[t].cullGroup = 0
		for _, groupName := range shapeMeshes[tile.ShapeID].GetGroupNames() {
			meshCullGroup := cullGroupFromName[groupName]
			//Transform the shape's culling group vector into the tile's space to assign the map space culling group
			dir := mgl32.Vec3{float32(meshCullGroup.dir[0]), float32(meshCullGroup.dir[1]), float32(meshCullGroup.dir[2])}
			dir = mgl32.TransformNormal(dir, rotMatx)
			intDir := [3]int{int(dir.X()), int(dir.Y()), int(dir.Z())}

			te3.Tiles.Data[t].cullGroup |= cullGroupFromDir[intDir].mask
		}
	}

	meshGroups := make([]Group, 0, len(groupTiles))
	meshGroupNames := make([]string, 0, len(groupTiles))

	//Add vertex data from tiles to map mesh
	for texID, tileIndices := range groupTiles {
		group := Group{Offset: len(mapInds), Length: 0}

		for _, t := range tileIndices {
			tile := te3.Tiles.Data[t]
			shapeMesh := shapeMeshes[tile.ShapeID]
			gridX, gridY, gridZ := te3.Tiles.GetGridPos(t)

			// tileVertexOffset := len(mapVerts.Pos)

			//Add the shape's groups individually to the aggregate mesh
			for _, shapeGroupName := range shapeMesh.GetGroupNames() {
				//Transform the culling direction of the shape by the tile rotation to get the right culling group
				g := cullGroupFromName[shapeGroupName]
				dir := mgl32.Vec3{float32(g.dir[0]), float32(g.dir[1]), float32(g.dir[2])}
				dir = mgl32.TransformNormal(dir, tile.rotMatx)
				intDir := [3]int{int(dir.X()), int(dir.Y()), int(dir.Z())}

				if te3.Tiles.ShouldCull(tile, gridX, gridY, gridZ, cullGroupFromDir[intDir]) {
					continue
				}

				shapeGroup := shapeMesh.GetGroup(shapeGroupName)

				for i := shapeGroup.Offset; i < shapeGroup.Offset+shapeGroup.Length; i++ {
					ind := shapeMesh.Inds[i]
					mapInds = append(mapInds, uint32(len(mapVerts.Pos)))

					//Add the shape's vertex position to the aggregate mesh, offset by the overall tile position
					pos := mgl32.TransformCoordinate(shapeMesh.Verts.Pos[ind], tile.rotMatx) //Rotate by tile orientation
					pos[0] += float32(gridX)*GRID_SPACING + HALF_GRID_SPACING
					pos[1] += float32(gridY)*GRID_SPACING + HALF_GRID_SPACING
					pos[2] += float32(gridZ)*GRID_SPACING + HALF_GRID_SPACING
					mapVerts.Pos = append(mapVerts.Pos, pos)

					//Append tex coordinates
					mapVerts.TexCoord = append(mapVerts.TexCoord, shapeMesh.Verts.TexCoord[ind])

					//Append normal, rotated by the tile orientation
					normal := mgl32.TransformNormal(shapeMesh.Verts.Normal[ind], tile.rotMatx)
					mapVerts.Normal = append(mapVerts.Normal, normal)
				}

				group.Length += shapeGroup.Length
			}
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
