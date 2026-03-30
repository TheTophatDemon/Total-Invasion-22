package comps

import (
	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3/mesh_gen"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

// Handles rendering and physics for a grid of tiles based on a TE3 map.
// Rendering geometry is optional.
type MapLayer struct {
	GridShape      collision.Grid
	Name           string
	tileAnims      []AnimationPlayer // Animates each texture group of tiles
	groupRenderers []MeshRender      // Renders each texture group of tiles
}

// Creates a map layer that renders geometry.
// `collisionLayer` is assigned to the map's GridShape.
// `excludeFlags` specifies texture flags that will be invisible.
func NewMapLayer(te3File *te3.TE3File, collisionLayer collision.Mask, excludeFlags []string) (MapLayer, error) {
	mesh, err := mesh_gen.BuildMeshFromTE3Map(te3File, excludeFlags)
	if err != nil {
		return MapLayer{}, err
	}
	cache.TakeMesh(te3File.FilePath(), mesh)

	gridShape := collision.NewGrid(te3File.Tiles.Width, te3File.Tiles.Height, te3File.Tiles.Length, te3.GridSpacing)
	gameMap := MapLayer{
		Name:      te3File.FilePath(),
		GridShape: gridShape,
	}
	gameMap.tileAnims = make([]AnimationPlayer, mesh.GroupCount())
	gameMap.groupRenderers = make([]MeshRender, mesh.GroupCount())

	for g, groupName := range mesh.GroupNames() {
		tex := cache.GetTexture(groupName)
		// Add animations if applicable
		if tex.IsAtlas() {
			anim, _ := tex.GetAnimation(tex.GetAnimationNames()[0])
			gameMap.tileAnims[g] = NewAnimationPlayer(anim, true)
		}

		// Add mesh component
		gameMap.groupRenderers[g] = NewMeshRenderGroup(mesh, shaders.MapShader, tex, groupName)
	}

	return gameMap, nil
}

func (gm *MapLayer) Update(deltaTime float32) {
	for i := range gm.tileAnims {
		gm.tileAnims[i].Update(deltaTime)
	}
}

func (gm *MapLayer) Render(context *render.Context) {
	for i := range gm.groupRenderers {
		gm.groupRenderers[i].Render(mgl32.Vec3{}, &gm.tileAnims[i], context)
	}
}
