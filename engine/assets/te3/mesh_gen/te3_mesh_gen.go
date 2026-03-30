package mesh_gen

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type cullInfo []math2.Triangle

var (
	normalEast  = mgl32.Vec3{1.0, 0.0, 0.0}
	normalWest  = mgl32.Vec3{-1.0, 0.0, 0.0}
	normalNorth = mgl32.Vec3{0.0, 0.0, -1.0}
	normalSouth = mgl32.Vec3{0.0, 0.0, 1.0}
	normalUp    = mgl32.Vec3{0.0, 1.0, 0.0}
	normalDown  = mgl32.Vec3{0.0, -1.0, 0.0}
)

// Transforms the triangle from a tile's shape mesh into the space of a tile.
func transformedTileTriangle(gridX, gridY, gridZ int, triangle math2.Triangle, rotation mgl32.Mat4) math2.Triangle {
	outTriangle := math2.Triangle{}
	for i := range len(triangle) {
		// Rotate and translate the points of the triangle to the tile's final position
		outTriangle[i] = mgl32.TransformCoordinate(triangle[i], rotation)
		outTriangle[i][0] += float32(gridX)*te3.GridSpacing + te3.HalfGridSpacing
		outTriangle[i][1] += float32(gridY)*te3.GridSpacing + te3.HalfGridSpacing
		outTriangle[i][2] += float32(gridZ)*te3.GridSpacing + te3.HalfGridSpacing
	}
	return outTriangle
}

// Returns true if the triangle happens to match with one from a neighboring tile.
func shouldCull(file *te3.TE3File, gridX, gridY, gridZ int, triangle math2.Triangle, tileCache []cullInfo) bool {
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

	if !file.Tiles.OutOfBounds(nborX, nborY, nborZ) {
		// Check the faces of the neighboring tile
		nborIdx := file.Tiles.FlattenGridPos(nborX, nborY, nborZ)
		nborTile := file.Tiles.Data[nborIdx]
		if nborTile.ShapeID < 0 {
			return false
		}

		// Otherwise, look for a triangle on the neighbor on such a plane that shares all three points with this triangle.
		for _, nborTriangle := range tileCache[nborIdx] {
			nborPlane := nborTriangle.Plane()

			//TODO: Might get more effective culling using a BSP based technique
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

// Creates a mesh from the tiles in the map. The result is not cached, so don't call this too often.
// The excludeTags parameter is used to provide a list of texture flags that will not generate any geometry.
func BuildMeshFromTE3Map(file *te3.TE3File, excludeFlags []string) (*geom.Mesh, error) {
	var err error

	// cpuProfile, err := os.Create("buildMesh.pprof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer cpuProfile.Close()
	// if err := pprof.StartCPUProfile(cpuProfile); err != nil {
	// 	log.Fatal(err)
	// }
	// defer pprof.StopCPUProfile()

	mapVerts := geom.Vertices{
		Pos:      make([]mgl32.Vec3, 0, len(file.Tiles.Data)*24),
		TexCoord: make([]mgl32.Vec2, 0, len(file.Tiles.Data)*24),
		Normal:   make([]mgl32.Vec3, 0, len(file.Tiles.Data)*24),
		Color:    nil,
	}
	mapInds := make([]uint32, 0, len(file.Tiles.Data)*12)

	shapeMeshes := make([]*geom.Mesh, len(file.Tiles.Shapes))
	for i, path := range file.Tiles.Shapes {
		shapeMeshes[i], err = cache.GetMesh(path)
		if err != nil {
			return nil, fmt.Errorf("shape mesh at %s not found", path)
		}
	}

	// Groups tile data indices by their texture
	groupTiles := make(map[te3.TextureID][]int, len(file.Tiles.Textures))

	// Find visible textures
textureLoop:
	for id, texPath := range file.Tiles.Textures {
		tex := cache.GetTexture(texPath)
		for _, flag := range excludeFlags {
			if tex == nil || tex.HasFlag(flag) {
				continue textureLoop
			}
		}
		groupTiles[te3.TextureID(id)] = make([]int, 0, 32)
	}

	// Caches the transformed triangles belonging to each tile.
	tileTriangles := make([]cullInfo, len(file.Tiles.Data))

	// Preprocess tiles
	for t, tile := range file.Tiles.Data {
		// Only visible tiles are processed here
		if tile.ShapeID < 0 {
			continue
		}

		// Assign to group(s) based on texture
		visible := false
		for _, texId := range tile.TextureIDs {
			if group, ok := groupTiles[texId]; ok {
				groupTiles[texId] = append(group, t)
				visible = true
			}
		}

		if visible {
			// Determine triangle orientations
			rotMatrix := tile.GetRotationMatrix()
			gridX, gridY, gridZ := file.Tiles.UnflattenGridPos(t)
			shapeTriIter := shapeMeshes[tile.ShapeID].IterTriangles()
			for shapeTriIter.HasNext() {
				tileTriangles[t] = append(tileTriangles[t], transformedTileTriangle(gridX, gridY, gridZ, shapeTriIter.Next(), rotMatrix))
			}
		}
	}

	meshGroups := make([]geom.Group, 0, len(groupTiles))
	meshGroupNames := make([]string, 0, len(groupTiles))

	// Add vertex data from tiles to map mesh
	for texID, tileIndices := range groupTiles {
		outGroup := geom.Group{Offset: len(mapInds), Length: 0}

		for _, ti := range tileIndices {
			tile := file.Tiles.Data[ti]
			shapeMesh := shapeMeshes[tile.ShapeID]
			gridX, gridY, gridZ := file.Tiles.UnflattenGridPos(ti)

			rotMatrix := tile.GetRotationMatrix()

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

			for tri := range len(shapeInds) / 3 {
				// Get triangle coordinates
				triangle := tileTriangles[ti][(shapeGroup.Offset/3)+tri]

				// Skip if culling tile
				if shouldCull(file, gridX, gridY, gridZ, triangle, tileTriangles) {
					continue
				}

				// Add the triangle's indices to the map mesh
				for i := range 3 {
					ind := shapeInds[(tri*3)+i]
					mapInds = append(mapInds, uint32(len(mapVerts.Pos)))

					// Vertices that are near grid boundaries are snapped to grid boundaries.
					// This covers up floating point errors that result in flickering bands of color on the screen.
					for v, coord := range triangle[i] {
						lowerBound := math2.Floor(coord/te3.GridSpacing) * te3.GridSpacing
						upperBound := math2.Ceil(coord/te3.GridSpacing) * te3.GridSpacing
						if coord-lowerBound < 0.05 {
							triangle[i][v] = lowerBound
						} else if upperBound-coord < 0.05 {
							triangle[i][v] = upperBound
						}
					}

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
		meshGroupNames = append(meshGroupNames, file.Tiles.Textures[texID])
	}

	mesh := geom.CreateMesh(mapVerts, mapInds)
	mesh.Upload()

	// Set group names to texture paths
	for g, group := range meshGroups {
		mesh.SetGroup(meshGroupNames[g], group)
	}

	return mesh, nil
}
