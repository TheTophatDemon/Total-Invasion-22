package te3

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

var (
	normalEast  = mgl32.Vec3{1.0, 0.0, 0.0}
	normalWest  = mgl32.Vec3{-1.0, 0.0, 0.0}
	normalNorth = mgl32.Vec3{0.0, 0.0, -1.0}
	normalSouth = mgl32.Vec3{0.0, 0.0, 1.0}
	normalUp    = mgl32.Vec3{0.0, 1.0, 0.0}
	normalDown  = mgl32.Vec3{0.0, -1.0, 0.0}
)

// Associates a tile index from a .te3 file with a set of triangles from the generated model.
type TriMap [][]math2.Triangle

// Transforms the triangle from a tile's shape mesh into the space of a tile.
func transformedTileTriangle(gridX, gridY, gridZ int, triangle math2.Triangle, rotation mgl32.Mat4) math2.Triangle {
	outTriangle := math2.Triangle{}
	for i := 0; i < len(triangle); i++ {
		// Rotate and translate the points of the triangle to the tile's final position
		outTriangle[i] = mgl32.TransformCoordinate(triangle[i], rotation)
		outTriangle[i][0] += float32(gridX)*GRID_SPACING + HALF_GRID_SPACING
		outTriangle[i][1] += float32(gridY)*GRID_SPACING + HALF_GRID_SPACING
		outTriangle[i][2] += float32(gridZ)*GRID_SPACING + HALF_GRID_SPACING
	}
	return outTriangle
}

// Returns true if the triangle happens to match with one from a neighboring tile.
func (te3 *TE3File) shouldCull(gridX, gridY, gridZ int, triangle math2.Triangle, shapeMeshes []*assets.Mesh) bool {
	plane := triangle.Plane()

	// Determine the grid position of the tile neighboring this face.
	nborX, nborY, nborZ := gridX, gridY, gridZ
	if plane.Normal.ApproxEqual(normalEast) {
		nborX += 1
	} else if plane.Normal.ApproxEqual(normalWest) {
		nborX -= 1
	} else if plane.Normal.ApproxEqual(normalNorth) {
		nborZ -= 1
	} else if plane.Normal.ApproxEqual(normalSouth) {
		nborZ += 1
	} else if plane.Normal.ApproxEqual(normalUp) {
		nborY += 1
	} else if plane.Normal.ApproxEqual(normalDown) {
		nborY -= 1
	} else {
		return false
	}

	if nborX >= 0 && nborY >= 0 && nborZ >= 0 && nborX < te3.Tiles.Width && nborY < te3.Tiles.Height && nborZ < te3.Tiles.Length {
		// Check the faces of the neighboring tile
		nborTile := te3.Tiles.Data[te3.Tiles.FlattenGridPos(nborX, nborY, nborZ)]
		if nborTile.ShapeID < 0 || assets.GetTexture(te3.Tiles.Textures[nborTile.TextureID]).HasFlag("invisible") {
			return false
		}
		nborRotation := nborTile.GetRotationMatrix()
		nborMesh := shapeMeshes[nborTile.ShapeID]
		for nt := 0; nt < len(nborMesh.Inds)/3; nt++ {
			nborTriangle := transformedTileTriangle(nborX, nborY, nborZ, math2.Triangle{
				nborMesh.Verts.Pos[nborMesh.Inds[nt*3+0]],
				nborMesh.Verts.Pos[nborMesh.Inds[nt*3+1]],
				nborMesh.Verts.Pos[nborMesh.Inds[nt*3+2]],
			}, nborRotation)
			nborPlane := nborTriangle.Plane()

			// The triangle is culled if the neighbor has a triangle facing the opposite direction sharing all three points.
			if mgl32.FloatEqual(mgl32.Abs(plane.Dist), mgl32.Abs(nborPlane.Dist)) &&
				mgl32.FloatEqual(nborPlane.Normal.Dot(plane.Normal), -1.0) &&
				(triangle[0].ApproxEqual(nborTriangle[0]) || triangle[0].ApproxEqual(nborTriangle[1]) || triangle[0].ApproxEqual(nborTriangle[2])) &&
				(triangle[1].ApproxEqual(nborTriangle[0]) || triangle[1].ApproxEqual(nborTriangle[1]) || triangle[1].ApproxEqual(nborTriangle[2])) &&
				(triangle[2].ApproxEqual(nborTriangle[0]) || triangle[2].ApproxEqual(nborTriangle[1]) || triangle[2].ApproxEqual(nborTriangle[2])) {
				return true
			}
		}
	}

	return false
}

func (te3 *TE3File) BuildMesh() (*assets.Mesh, TriMap, error) {
	var err error

	mapVerts := assets.Vertices{
		Pos:      make([]mgl32.Vec3, 0),
		TexCoord: make([]mgl32.Vec2, 0),
		Normal:   make([]mgl32.Vec3, 0),
		Color:    nil,
	}
	mapInds := make([]uint32, 0)

	shapeMeshes := make([]*assets.Mesh, len(te3.Tiles.Shapes))
	for i, path := range te3.Tiles.Shapes {
		shapeMeshes[i], err = assets.GetMesh(path)
		if err != nil {
			return nil, nil, fmt.Errorf("shape mesh at %s not found", path)
		}
	}

	// Groups tile data indices by their texture
	groupTiles := make(map[TextureID][]int)

	// Preprocess tiles
	for t, tile := range te3.Tiles.Data {
		// Only visible tiles are processed here
		if tile.ShapeID < 0 || tile.TextureID < 0 {
			continue
		}

		// Exclude tiles with invisible texture flag
		if assets.GetTexture(te3.Tiles.Textures[tile.TextureID]).HasFlag("invisible") {
			continue
		}

		// Assign to group based on texture
		group, ok := groupTiles[tile.TextureID]
		if !ok {
			group = make([]int, 0, 16)
		}
		groupTiles[tile.TextureID] = append(group, t)
	}

	meshGroups := make([]assets.Group, 0, len(groupTiles))
	meshGroupNames := make([]string, 0, len(groupTiles))

	triMap := make(TriMap, len(te3.Tiles.Data))

	// Add vertex data from tiles to map mesh
	for texID, tileIndices := range groupTiles {
		group := assets.Group{Offset: len(mapInds), Length: 0}

		for _, ti := range tileIndices {
			tile := te3.Tiles.Data[ti]
			shapeMesh := shapeMeshes[tile.ShapeID]
			gridX, gridY, gridZ := te3.Tiles.GetGridPos(ti)

			rotMatrix := tile.GetRotationMatrix()

			// Create triangle map array for this tile
			triMap[ti] = make([]math2.Triangle, 0, 8)

			for tri := 0; tri < len(shapeMesh.Inds)/3; tri++ {
				// Get triangle coordinates
				triangle := transformedTileTriangle(gridX, gridY, gridZ, math2.Triangle{
					shapeMesh.Verts.Pos[shapeMesh.Inds[tri*3+0]],
					shapeMesh.Verts.Pos[shapeMesh.Inds[tri*3+1]],
					shapeMesh.Verts.Pos[shapeMesh.Inds[tri*3+2]],
				}, rotMatrix)

				// Add to triangle map
				triMap[ti] = append(triMap[ti], triangle)

				// Skip if culling tile
				if te3.shouldCull(gridX, gridY, gridZ, triangle, shapeMeshes) {
					continue
				}

				// Add the triangle's indices to the map mesh
				for i := 0; i < 3; i++ {
					ind := shapeMesh.Inds[tri*3+i]
					mapInds = append(mapInds, uint32(len(mapVerts.Pos)))

					// Add the shape's vertex position to the aggregate mesh, offset by the overall tile position
					mapVerts.Pos = append(mapVerts.Pos, triangle[i])

					// Append tex coordinates
					mapVerts.TexCoord = append(mapVerts.TexCoord, shapeMesh.Verts.TexCoord[ind])

					// Append normal, rotated by the tile orientation
					normal := mgl32.TransformNormal(shapeMesh.Verts.Normal[ind], rotMatrix)
					mapVerts.Normal = append(mapVerts.Normal, normal)
				}
				group.Length += 3
			}
		}

		meshGroups = append(meshGroups, group)
		meshGroupNames = append(meshGroupNames, te3.Tiles.Textures[texID])
	}

	mesh := assets.CreateMesh(mapVerts, mapInds)

	// Set group names to texture paths
	for g, group := range meshGroups {
		mesh.SetGroup(meshGroupNames[g], group)
	}

	return mesh, triMap, nil
}
