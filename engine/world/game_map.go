package world

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
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

	// Check for triangle hits in the center first, then the edges.
	// This prevents triangle edges from stopping smooth movement along neighboring triangles.
	filter := collision.TRIHIT_CENTER
	for {
		// Iterate over the subset of tiles that the body occupies
		bbox := body.Shape.Extents().Translate(body.Transform.Position())
		i, j, k := gm.tiles.WorldToGridPos(bbox.Max)
		l, m, n := gm.tiles.WorldToGridPos(bbox.Min)
		minX, minY, minZ := max(0, min(i, l)), max(0, min(j, m)), max(0, min(k, n))
		maxX, maxY, maxZ := min(max(i, l), gm.tiles.Width-1), min(max(j, m), gm.tiles.Height-1), min(max(k, n), gm.tiles.Length-1)
		for x := minX; x <= maxX; x += 1 {
			for y := minY; y <= maxY; y += 1 {
				for z := minZ; z <= maxZ; z += 1 {
					t := gm.tiles.FlattenGridPos(x, y, z)
					if gm.tiles.Data[t].ShapeID >= 0 {
						body.ResolveCollisionSphereTriangles(mgl32.Vec3{}, gm.mesh, gm.triMap[t], filter)
					}
				}
			}
		}

		if filter == collision.TRIHIT_CENTER {
			filter = collision.TRIHIT_EDGE
		} else {
			break
		}
	}

	return nil
}
