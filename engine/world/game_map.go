package world

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/containers"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Map struct {
	tiles          te3.Tiles
	mesh           *geom.Mesh
	triMap         te3.TriMap              // Maps a flattened tile index to its indices in the mesh's triangles array.
	shapeMap       []collision.Shape       // Maps a tile's index to its collision shape.
	tileAnims      []comps.AnimationPlayer // Animates each texture group of tiles
	groupRenderers []comps.MeshRender      // Renders each texture group of tiles
}

func NewMap(te3File *te3.TE3File) (*Map, error) {
	mesh, triMap, err := te3File.BuildMesh()
	if err != nil {
		return nil, err
	}
	cache.TakeMesh(te3File.FilePath(), mesh)

	// Set all tile collision shapes to use the triangle mesh by default.
	shapeMap := make([]collision.Shape, len(te3File.Tiles.Data))
	for s := range shapeMap {
		if te3File.Tiles.Data[s].ShapeID >= 0 {
			shapeMap[s] = collision.NewMeshSubset(mesh, triMap[s])
		} else {
			shapeMap[s] = nil
		}
	}

	gameMap := &Map{
		tiles:          te3File.Tiles,
		mesh:           mesh,
		triMap:         triMap,
		shapeMap:       shapeMap,
		tileAnims:      make([]comps.AnimationPlayer, mesh.GroupCount()),
		groupRenderers: make([]comps.MeshRender, mesh.GroupCount()),
	}

	for g, groupName := range mesh.GroupNames() {
		tex := cache.GetTexture(groupName)
		// Add animations if applicable
		if tex.IsAtlas() {
			anim, _ := tex.GetAnimation(tex.GetAnimationNames()[0])
			gameMap.tileAnims[g] = comps.NewAnimationPlayer(anim, true)
		}

		// Add mesh component
		gameMap.groupRenderers[g] = comps.NewMeshRenderGroup(mesh, shaders.MapShader, tex, groupName)
	}

	return gameMap, nil
}

func (gm *Map) SetTileCollisionShapes(shapeName string, shape collision.Shape) error {
	return gm.SetTileCollisionShapesForAngles(shapeName, 0, 360, 0, 360, shape)
}

// Sets the collision shape of all tiles that have the specified shape, and whose angles are within the designated ranges.
// 'yawMin' and 'pitchMin' are inclusive bounds, but 'yawMax' and 'pitchMax' are exclusive.
func (gm *Map) SetTileCollisionShapesForAngles(shapeName string, yawMin, yawMax, pitchMin, pitchMax int32, shape collision.Shape) error {
	var shapeID te3.ShapeID = -1
	for id, name := range gm.tiles.Shapes {
		if name == shapeName {
			shapeID = te3.ShapeID(id)
			break
		}
	}
	if shapeID < 0 {
		return fmt.Errorf("shape not found")
	}
	for index, tile := range gm.tiles.Data {
		if tile.ShapeID == shapeID &&
			tile.Yaw >= yawMin && tile.Yaw < yawMax &&
			tile.Pitch >= pitchMin && tile.Pitch < pitchMax {

			gm.shapeMap[index] = shape
		}
	}
	return nil
}

func (gm *Map) Update(deltaTime float32) {
	for i := range gm.tileAnims {
		gm.tileAnims[i].Update(deltaTime)
	}
}

func (gm *Map) Render(context *render.Context) {
	for i := range gm.groupRenderers {
		gm.groupRenderers[i].Render(nil, &gm.tileAnims[i], context)
	}
}

func (gm *Map) CastRay(rayOrigin, rayDir mgl32.Vec3) collision.RaycastResult {
	if lenSqr := rayDir.LenSqr(); lenSqr == 0.0 {
		return collision.RaycastResult{}
	} else if lenSqr != 1.0 {
		rayDir = rayDir.Normalize()
	}

	pos := rayOrigin
	for {
		i, j, k := gm.tiles.WorldToGridPos(pos)

		if gm.tiles.OutOfBounds(i, j, k) {
			break
		}

		// Find positions on the XZ, XY, and YZ planes that will be hit next by the ray.
		offsetFromCorner := pos.Sub(gm.tiles.GridToWorldPos(i, j, k, false))
		if offsetFromCorner.Z() == 0.0 && rayDir.Z() < 0.0 {
			k -= 1
		}
		if offsetFromCorner.Y() == 0.0 && rayDir.Y() < 0.0 {
			j -= 1
		}
		if offsetFromCorner.X() == 0.0 && rayDir.X() < 0.0 {
			i -= 1
		}
		offsetFromCorner = pos.Sub(gm.tiles.GridToWorldPos(i, j, k, false))

		var nextYZPlane, nextXZPlane, nextXYPlane mgl32.Vec3
		var smallestDist float32 = math.MaxFloat32
		var nextPos mgl32.Vec3

		// Find next YZ plane
		var skip bool
		if rayDir.X() > 0.0 {
			nextYZPlane = mgl32.Vec3{pos.X() - offsetFromCorner.X() + gm.tiles.GridSpacing()}
		} else if rayDir.X() < 0.0 {
			nextYZPlane = mgl32.Vec3{pos.X() - offsetFromCorner.X()}
		} else {
			skip = true
		}
		if !skip {
			t := (nextYZPlane.X() - pos.X()) / rayDir.X()
			nextYZPlane = mgl32.Vec3{nextYZPlane.X(), pos.Y() + rayDir.Y()*t, pos.Z() + rayDir.Z()*t}
			if t < smallestDist {
				smallestDist = t
				nextPos = nextYZPlane
			}
		}

		// Find next XZ plane
		skip = false
		if rayDir.Y() > 0.0 {
			nextXZPlane = mgl32.Vec3{0.0, pos.Y() - offsetFromCorner.Y() + gm.tiles.GridSpacing()}
		} else if rayDir.Y() < 0.0 {
			nextXZPlane = mgl32.Vec3{0.0, pos.Y() - offsetFromCorner.Y()}
		} else {
			skip = true
		}
		if !skip {
			t := (nextXZPlane.Y() - pos.Y()) / rayDir.Y()
			nextXZPlane = mgl32.Vec3{pos.X() + rayDir.X()*t, nextXZPlane.Y(), pos.Z() + rayDir.Z()*t}
			if t < smallestDist {
				smallestDist = t
				nextPos = nextXZPlane
			}
		}

		// Find next XY plane
		skip = false
		if rayDir.Z() > 0.0 {
			nextXYPlane = mgl32.Vec3{0.0, 0.0, pos.Z() - offsetFromCorner.Z() + gm.tiles.GridSpacing()}
		} else if rayDir.Z() < 0.0 {
			nextXYPlane = mgl32.Vec3{0.0, 0.0, pos.Z() - offsetFromCorner.Z()}
		} else {
			skip = true
		}
		if !skip {
			t := (nextXYPlane.Z() - pos.Z()) / rayDir.Z()
			nextXYPlane = mgl32.Vec3{pos.X() + rayDir.X()*t, pos.Y() + rayDir.Y()*t, nextXYPlane.Z()}
			if t < smallestDist {
				smallestDist = t
				nextPos = nextXYPlane
			}
		}

		// Check for a tile at that grid cell
		i, j, k = gm.tiles.WorldToGridPos(nextPos)
		if nextPos.ApproxEqual(nextXYPlane) && rayDir.Z() < 0.0 {
			k -= 1
		} else if nextPos.ApproxEqual(nextXZPlane) && rayDir.Y() < 0.0 {
			j -= 1
		} else if nextPos.ApproxEqual(nextYZPlane) && rayDir.X() < 0.0 {
			i -= 1
		}
		if gm.tiles.OutOfBounds(i, j, k) {
			break
		}

		tileIndex := gm.tiles.FlattenGridPos(i, j, k)
		if tileShape := gm.shapeMap[tileIndex]; tileShape != nil {
			// Exclude tiles with invisible texture flag
			if cache.GetTexture(gm.tiles.Textures[gm.tiles.Data[tileIndex].TextureID]).HasFlag("invisible") {
				goto next
			}

			rayDir := nextPos.Sub(pos)
			if rayDir.LenSqr() <= 0.0 {
				break
			}
			rayDir = rayDir.Normalize()
			tileCenter := gm.tiles.GridToWorldPos(i, j, k, true)
			if cast := tileShape.Raycast(pos, rayDir, tileCenter); cast.Hit {
				return cast
			}
		}

	next:
		pos = nextPos
	}

	return collision.RaycastResult{}
}

// Moves the body in response to collisions with the tiles in this game map.
func (gm *Map) ResolveCollision(body *comps.Body) error {
	if body.NoClip || body.Velocity.LenSqr() == 0.0 {
		return nil
	}

	if _, isSphere := body.Shape.(collision.Sphere); !isSphere {
		return fmt.Errorf("body must have a sphere collision shape")
	}

	// Iterate over the subset of tiles that the body occupies
	bbox := body.Shape.Extents().Translate(body.Transform.Position())
	i, j, k := gm.tiles.WorldToGridPos(bbox.Max)
	l, m, n := gm.tiles.WorldToGridPos(bbox.Min)
	minX, minY, minZ := max(0, min(i, l)), max(0, min(j, m)), max(0, min(k, n))
	maxX, maxY, maxZ := min(max(i, l), gm.tiles.Width-1), min(max(j, m), gm.tiles.Height-1), min(max(k, n), gm.tiles.Length-1)
	tileCount := (maxX - minX + 1) * (maxZ - minZ + 1) * (maxY - minY + 1)
	if tileCount <= 0 {
		return nil
	}

	// Visit each tile within the movement range, starting from the one closest to the current position and then proceeding to its neighbors.
	isChecked := make(map[[3]int]bool)
	visitQueue := containers.NewRingBuffer(make([][3]int, tileCount))
	startX, startY, startZ := gm.tiles.WorldToGridPos(body.Transform.Position())
	if gm.tiles.OutOfBounds(startX, startY, startZ) {
		return nil
	}
	start := [3]int{startX, startY, startZ}
	visitQueue.Enqueue(start)
	isChecked[start] = true

	for pos, empty := visitQueue.Dequeue(); !empty; pos, empty = visitQueue.Dequeue() {
		t := gm.tiles.FlattenGridPos(pos[0], pos[1], pos[2])
		if gm.tiles.Data[t].ShapeID >= 0 {
			// Resolve collision against this tile
			tileCenter := gm.tiles.GridToWorldPos(pos[0], pos[1], pos[2], true)
			tileShape := gm.shapeMap[t]
			switch shape := tileShape.(type) {
			case collision.Sphere:
				_ = body.ResolveCollisionSphereSphere(tileCenter, shape.Radius(), 1.0)
			case collision.Box:
				_ = body.ResolveCollisionSphereBox(tileCenter, tileShape.Extents(), 1.0)
			case collision.Mesh:
				// Check for triangle hits in the center first, then the edges.
				// This prevents triangle edges from stopping smooth movement along neighboring triangles.
				body.ResolveCollisionSphereTriangles(mgl32.Vec3{}, gm.mesh, gm.triMap[t], collision.TRIHIT_CENTER)
				body.ResolveCollisionSphereTriangles(mgl32.Vec3{}, gm.mesh, gm.triMap[t], collision.TRIHIT_ALL)
			}
		}

		// Add neighboring tiles to the queue
		neighbors := [...][3]int{
			{pos[0] + 1, pos[1], pos[2]},
			{pos[0] - 1, pos[1], pos[2]},
			{pos[0], pos[1] + 1, pos[2]},
			{pos[0], pos[1] - 1, pos[2]},
			{pos[0], pos[1], pos[2] + 1},
			{pos[0], pos[1], pos[2] - 1},
		}
		for _, n := range neighbors {
			if n[0] < minX || n[1] < minY || n[2] < minZ || n[0] > maxX || n[1] > maxY || n[2] > maxZ {
				continue
			}
			_, v := isChecked[n]
			if !v {
				visitQueue.Enqueue(n)
				isChecked[n] = true
			}
		}
	}

	return nil
}
