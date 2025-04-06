package te3

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
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

// Associates a tile index from a .te3 file with a set of triangle indices from the generated model.
type TriMap [][]int

// Transforms the triangle from a tile's shape mesh into the space of a tile.
func transformedTileTriangle(gridX, gridY, gridZ int, triangle math2.Triangle, rotation mgl32.Mat4) math2.Triangle {
	outTriangle := math2.Triangle{}
	for i := range len(triangle) {
		// Rotate and translate the points of the triangle to the tile's final position
		outTriangle[i] = mgl32.TransformCoordinate(triangle[i], rotation)
		outTriangle[i][0] += float32(gridX)*GRID_SPACING + HALF_GRID_SPACING
		outTriangle[i][1] += float32(gridY)*GRID_SPACING + HALF_GRID_SPACING
		outTriangle[i][2] += float32(gridZ)*GRID_SPACING + HALF_GRID_SPACING
	}
	return outTriangle
}

// Returns true if the triangle happens to match with one from a neighboring tile.
func (te3 *TE3File) shouldCull(gridX, gridY, gridZ int, triangle math2.Triangle, shapeMeshes []*geom.Mesh) bool {
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
		if nborTile.ShapeID < 0 {
			return false
		}
		nborRotation := nborTile.GetRotationMatrix()
		nborMesh := shapeMeshes[nborTile.ShapeID]
		for iter := nborMesh.IterTriangles(); iter.HasNext(); {
			nborTri := iter.Next()
			nborTriangle := transformedTileTriangle(nborX, nborY, nborZ, nborTri, nborRotation)
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

func (te3 *TE3File) BuildMesh() (*geom.Mesh, TriMap, error) {
	var err error

	mapVerts := geom.Vertices{
		Pos:      make([]mgl32.Vec3, 0, len(te3.Tiles.Data)*24),
		TexCoord: make([]mgl32.Vec2, 0, len(te3.Tiles.Data)*24),
		Normal:   make([]mgl32.Vec3, 0, len(te3.Tiles.Data)*24),
		Color:    nil,
	}
	mapInds := make([]uint32, 0, len(te3.Tiles.Data)*12)

	shapeMeshes := make([]*geom.Mesh, len(te3.Tiles.Shapes))
	for i, path := range te3.Tiles.Shapes {
		shapeMeshes[i], err = cache.GetMesh(path)
		if err != nil {
			return nil, nil, fmt.Errorf("shape mesh at %s not found", path)
		}
	}

	// Groups tile data indices by their texture
	groupTiles := make(map[TextureID][]int, len(te3.Tiles.Textures))

	// Preprocess tiles
	for t, tile := range te3.Tiles.Data {
		// Only visible tiles are processed here
		if tile.ShapeID < 0 {
			continue
		}

		// Assign to group(s) based on texture
		for _, texId := range tile.TextureIDs {
			group, ok := groupTiles[texId]
			if !ok {
				group = make([]int, 0, 16)
			}
			groupTiles[texId] = append(group, t)
		}
	}

	meshGroups := make([]geom.Group, 0, len(groupTiles))
	meshGroupNames := make([]string, 0, len(groupTiles))

	triMap := make(TriMap, len(te3.Tiles.Data))

	// Add vertex data from tiles to map mesh
	for texID, tileIndices := range groupTiles {
		outGroup := geom.Group{Offset: len(mapInds), Length: 0}

		for _, ti := range tileIndices {
			tile := te3.Tiles.Data[ti]
			shapeMesh := shapeMeshes[tile.ShapeID]
			gridX, gridY, gridZ := te3.Tiles.UnflattenGridPos(ti)

			rotMatrix := tile.GetRotationMatrix()

			// Create triangle map array for this tile
			triMap[ti] = make([]int, 0, 8)

			// Pick the material on the mesh used for this texture
			var shapeGroup geom.Group
			switch texID {
			case tile.TextureIDs[0]:
				shapeGroup = shapeMesh.Group("primary")
			case tile.TextureIDs[1]:
				shapeGroup = shapeMesh.Group("secondary")
			}
			shapeInds := shapeMesh.Inds()
			if tile.TextureIDs[0] == tile.TextureIDs[1] {
				// Both textures are the same, so use the whole mesh.
				shapeGroup = geom.Group{}
			} else if shapeGroup != (geom.Group{}) {
				shapeInds = shapeInds[shapeGroup.Offset:][:shapeGroup.Length]
			}

			shapeTriIter := shapeMesh.IterTriangles()
			for range shapeGroup.Offset / 3 {
				shapeTriIter.Next()
			}

			for tri := range len(shapeInds) / 3 {
				// Get triangle coordinates
				triangle := transformedTileTriangle(gridX, gridY, gridZ, shapeTriIter.Next(), rotMatrix)

				// Skip if culling tile
				if te3.shouldCull(gridX, gridY, gridZ, triangle, shapeMeshes) {
					continue
				}

				// Add to triangle map
				triMap[ti] = append(triMap[ti], len(mapInds)/3)

				// Add the triangle's indices to the map mesh
				for i := range 3 {
					ind := shapeInds[(tri*3)+i]
					mapInds = append(mapInds, uint32(len(mapVerts.Pos)))

					// Add the shape's vertex position to the aggregate mesh, offset by the overall tile position
					mapVerts.Pos = append(mapVerts.Pos, triangle[i])

					// Append tex coordinates
					mapVerts.TexCoord = append(mapVerts.TexCoord, shapeMesh.Verts().TexCoord[ind])

					// Append normal, rotated by the tile orientation
					normal := mgl32.TransformNormal(shapeMesh.Verts().Normal[ind], rotMatrix)
					mapVerts.Normal = append(mapVerts.Normal, normal)
				}
				outGroup.Length += 3
			}
		}

		meshGroups = append(meshGroups, outGroup)
		meshGroupNames = append(meshGroupNames, te3.Tiles.Textures[texID])
	}

	mesh := geom.CreateMesh(mapVerts, mapInds)
	mesh.Upload()

	// Set group names to texture paths
	for g, group := range meshGroups {
		mesh.SetGroup(meshGroupNames[g], group)
	}

	return mesh, triMap, nil
}
