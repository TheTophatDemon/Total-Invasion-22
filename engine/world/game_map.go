package world

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Map struct {
	tiles          te3.Tiles
	mesh           *assets.Mesh
	triMap         te3.TriMap
	tileAnims      []comps.AnimationPlayer
	groupRenderers []comps.MeshRender
}

func NewMap(te3File *te3.TE3File) (*Map, error) {
	mesh, triMap, err := te3File.BuildMesh()
	if err != nil {
		return nil, err
	}
	assets.TakeMesh(te3File.FilePath(), mesh)

	gameMap := &Map{
		tiles:          te3File.Tiles,
		mesh:           mesh,
		triMap:         triMap,
		tileAnims:      make([]comps.AnimationPlayer, mesh.GetGroupCount()),
		groupRenderers: make([]comps.MeshRender, mesh.GetGroupCount()),
	}

	for g, groupName := range mesh.GetGroupNames() {
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
func (gm *Map) ResolveCollision(body *comps.Body) {
	// Iterate over the subset of tiles that the body occupies
	i, j, k := gm.tiles.WorldToGridPos(body.Transform.Position().Add(body.Extents))
	l, m, n := gm.tiles.WorldToGridPos(body.Transform.Position().Sub(body.Extents))
	minX, minY, minZ := min(i, l), min(j, m), min(k, n)
	maxX, maxY, maxZ := max(i, l), max(j, m), max(k, n)
	for x := minX; x <= maxX; x += 1 {
		for y := minY; y <= maxY; y += 1 {
			for z := minZ; z <= maxZ; z += 1 {
				t := gm.tiles.FlattenGridPos(x, y, z)
				if gm.tiles.Data[t].ShapeID >= 0 {
					body.ResolveCollisionTriangles(mgl32.Vec3{}, gm.triMap[t])
				}
			}
		}
	}
}
