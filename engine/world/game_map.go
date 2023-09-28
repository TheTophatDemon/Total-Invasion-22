package world

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/containers"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Map struct {
	tiles          te3.Tiles
	mesh           *assets.Mesh
	triMap         te3.TriMap              // Maps a flattened tile index to its indices in the mesh's triangles array.
	shapeMap       []collision.Shape       // Maps a tile's ShapeID to its collision shape.
	tileAnims      []comps.AnimationPlayer // Animates each texture group of tiles
	groupRenderers []comps.MeshRender      // Renders each texture group of tiles
}

func NewMap(te3File *te3.TE3File) (*Map, error) {
	mesh, triMap, err := te3File.BuildMesh()
	if err != nil {
		return nil, err
	}
	assets.TakeMesh(te3File.FilePath(), mesh)

	// Set all tile collision shapes to use the triangle mesh by default.
	shapeMap := make([]collision.Shape, len(te3File.Tiles.Shapes))
	for s := range shapeMap {
		shapeMap[s] = collision.ShapeMesh(mesh)
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
		tex := assets.GetTexture(groupName)
		// Add animations if applicable
		if tex.AnimationCount() > 0 {
			gameMap.tileAnims[g] = comps.NewAnimationPlayer(tex.GetAnimation(0), true)
		}

		// Add mesh component
		gameMap.groupRenderers[g] = comps.NewMeshRenderGroup(mesh, shaders.MapShader, tex, groupName)
	}

	return gameMap, nil
}

func (gm *Map) SetTileCollisionShape(shapeName string, shape collision.Shape) error {
	for id, name := range gm.tiles.Shapes {
		if name == shapeName {
			gm.shapeMap[id] = shape
			return nil
		}
	}
	return fmt.Errorf("shape not found")
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

// Moves the body in response to collisions with the tiles in this game map.
func (gm *Map) ResolveCollision(body *comps.Body) error {
	_, isSphere := body.Shape.Radius()
	if !isSphere {
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
		if shapeID := gm.tiles.Data[t].ShapeID; shapeID >= 0 {
			// Resolve collision against this tile
			tileCenter := gm.tiles.GridToWorldPos(pos[0], pos[1], pos[2], true)
			switch gm.shapeMap[shapeID].Kind() {
			case collision.SHAPE_KIND_SPHERE:
				tileRadius, _ := gm.shapeMap[shapeID].Radius()
				_ = body.ResolveCollisionSphereSphere(tileCenter, tileRadius)
			case collision.SHAPE_KIND_BOX:
				_ = body.ResolveCollisionSphereBox(tileCenter, gm.shapeMap[shapeID].Extents(), true)
			case collision.SHAPE_KIND_MESH:
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

	// filter := collision.TRIHIT_CENTER
	// for {
	// 	for x := minX; x <= maxX; x += 1 {
	// 		for y := minY; y <= maxY; y += 1 {
	// 			for z := minZ; z <= maxZ; z += 1 {
	// 				t := gm.tiles.FlattenGridPos(x, y, z)
	// 				if shapeID := gm.tiles.Data[t].ShapeID; shapeID >= 0 {
	// 					// Resolve collision against this tile
	// 					tileCenter := gm.tiles.GridToWorldPos(x, y, z, true)
	// 					switch gm.shapeMap[shapeID].Kind() {
	// 					case collision.SHAPE_KIND_SPHERE:
	// 						tileRadius, _ := gm.shapeMap[shapeID].Radius()
	// 						_ = body.ResolveCollisionSphereSphere(tileCenter, tileRadius)
	// 					case collision.SHAPE_KIND_BOX:
	// 						_ = body.ResolveCollisionSphereBox(tileCenter, gm.shapeMap[shapeID].Extents(), true)
	// 					case collision.SHAPE_KIND_MESH:
	// 						// Check for triangle hits in the center first, then the edges.
	// 						// This prevents triangle edges from stopping smooth movement along neighboring triangles.
	// 						body.ResolveCollisionSphereTriangles(mgl32.Vec3{}, gm.mesh, gm.triMap[t], collision.TRIHIT_CENTER)
	// 						body.ResolveCollisionSphereTriangles(mgl32.Vec3{}, gm.mesh, gm.triMap[t], collision.TRIHIT_ALL)
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}

	// 	if filter == collision.TRIHIT_CENTER {
	// 		filter = collision.TRIHIT_EDGE
	// 	} else {
	// 		break
	// 	}
	// }

	return nil
}
