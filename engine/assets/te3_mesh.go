package assets

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
)

var NORMAL_EAST = mgl32.Vec3{1.0, 0.0, 0.0}
var NORMAL_WEST = mgl32.Vec3{-1.0, 0.0, 0.0}
var NORMAL_NORTH = mgl32.Vec3{0.0, 0.0, -1.0}
var NORMAL_SOUTH = mgl32.Vec3{0.0, 0.0, 1.0}
var NORMAL_UP = mgl32.Vec3{0.0, 1.0, 0.0}
var NORMAL_DOWN = mgl32.Vec3{0.0, -1.0, 0.0}

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

			rotMatrix := tile.GetRotationMatrix()

		triangle:
			for t := 0; t < len(shapeMesh.Inds)/3; t++ {
				//Cull the triangle if it happens to match with one from a neighboring tile

				triPoints := [3]mgl32.Vec3{}
				for i := 0; i < 3; i++ {
					//Rotate and translate the points of the triangle to the tile's final position
					triPoints[i] = mgl32.TransformCoordinate(shapeMesh.Verts.Pos[shapeMesh.Inds[t*3+i]], rotMatrix)
					triPoints[i][0] += float32(gridX)*GRID_SPACING + HALF_GRID_SPACING
					triPoints[i][1] += float32(gridY)*GRID_SPACING + HALF_GRID_SPACING
					triPoints[i][2] += float32(gridZ)*GRID_SPACING + HALF_GRID_SPACING
				}
				planeNormal := triPoints[1].Sub(triPoints[0]).Cross(triPoints[2].Sub(triPoints[0])).Normalize()
				planeDist := -planeNormal.Dot(triPoints[0])

				//Determine the grid position of the tile neighboring this face
				nborX, nborY, nborZ := gridX, gridY, gridZ
				if planeNormal.ApproxEqual(NORMAL_EAST) {
					nborX += 1
				} else if planeNormal.ApproxEqual(NORMAL_WEST) {
					nborX -= 1
				} else if planeNormal.ApproxEqual(NORMAL_NORTH) {
					nborZ -= 1
				} else if planeNormal.ApproxEqual(NORMAL_SOUTH) {
					nborZ += 1
				} else if planeNormal.ApproxEqual(NORMAL_UP) {
					nborY += 1
				} else if planeNormal.ApproxEqual(NORMAL_DOWN) {
					nborY -= 1
				} else {
					goto nocull
				}

				if nborX >= 0 && nborY >= 0 && nborZ >= 0 && nborX < te3.Tiles.Width && nborY < te3.Tiles.Height && nborZ < te3.Tiles.Length {
					//Check the faces of the neighboring tile
					nborTile := te3.Tiles.Data[te3.Tiles.FlattenGridPos(nborX, nborY, nborZ)]
					if nborTile.ShapeID < 0 || GetTexture(te3.Tiles.Textures[nborTile.TextureID]).HasFlag("invisible") {
						goto nocull
					}
					nRotMatrix := nborTile.GetRotationMatrix()
					nborMesh := shapeMeshes[nborTile.ShapeID]
					for nt := 0; nt < len(nborMesh.Inds)/3; nt++ {
						nTriPoints := [3]mgl32.Vec3{}
						for i := 0; i < 3; i++ {
							//Rotate and translate the points of the triangle to the tile's final position
							nTriPoints[i] = mgl32.TransformCoordinate(nborMesh.Verts.Pos[nborMesh.Inds[nt*3+i]], nRotMatrix)
							nTriPoints[i][0] += float32(nborX)*GRID_SPACING + HALF_GRID_SPACING
							nTriPoints[i][1] += float32(nborY)*GRID_SPACING + HALF_GRID_SPACING
							nTriPoints[i][2] += float32(nborZ)*GRID_SPACING + HALF_GRID_SPACING
						}
						nPlaneNormal := nTriPoints[1].Sub(nTriPoints[0]).Cross(nTriPoints[2].Sub(nTriPoints[0])).Normalize()
						nPlaneDist := -nPlaneNormal.Dot(nTriPoints[0])

						//The triangle is culled if the neighbor has a triangle facing the opposite direction sharing all three points.
						if mgl32.FloatEqual(mgl32.Abs(planeDist), mgl32.Abs(nPlaneDist)) &&
							mgl32.FloatEqual(nPlaneNormal.Dot(planeNormal), -1.0) &&
							(triPoints[0].ApproxEqual(nTriPoints[0]) || triPoints[0].ApproxEqual(nTriPoints[1]) || triPoints[0].ApproxEqual(nTriPoints[2])) &&
							(triPoints[1].ApproxEqual(nTriPoints[0]) || triPoints[1].ApproxEqual(nTriPoints[1]) || triPoints[1].ApproxEqual(nTriPoints[2])) &&
							(triPoints[2].ApproxEqual(nTriPoints[0]) || triPoints[2].ApproxEqual(nTriPoints[1]) || triPoints[2].ApproxEqual(nTriPoints[2])) {
							continue triangle
						}
					}
				}

			nocull:
				//Add the triangle's indices to the map mesh
				for i := 0; i < 3; i++ {
					ind := shapeMesh.Inds[t*3+i]
					mapInds = append(mapInds, uint32(len(mapVerts.Pos)))

					//Add the shape's vertex position to the aggregate mesh, offset by the overall tile position
					mapVerts.Pos = append(mapVerts.Pos, triPoints[i])

					//Append tex coordinates
					mapVerts.TexCoord = append(mapVerts.TexCoord, shapeMesh.Verts.TexCoord[ind])

					//Append normal, rotated by the tile orientation
					normal := mgl32.TransformNormal(shapeMesh.Verts.Normal[ind], rotMatrix)
					mapVerts.Normal = append(mapVerts.Normal, normal)
				}
				group.Length += 3
			}
		}

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
