package comps

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

// Handles rendering and physics for a grid of tiles based on a TE3 map.
// Rendering geometry is optional.
type MapLayer struct {
	GridShape      collision.Grid
	Layer          collision.Mask
	Name           string
	tileAnims      []AnimationPlayer // Animates each texture group of tiles
	groupRenderers []MeshRender      // Renders each texture group of tiles
}

// Creates a map layer that renders geometry.
// `collisionLayer` is assigned to the map's GridShape.
// `excludeFlags` specifies texture flags that will be invisible.
func NewMainMapLayer(te3File *te3.TE3File, collisionLayer collision.Mask, excludeFlags []string) (MapLayer, error) {
	mesh, err := te3File.BuildMesh(excludeFlags)
	if err != nil {
		return MapLayer{}, err
	}
	cache.TakeMesh(te3File.FilePath(), mesh)

	gameMap := NewExtraMapLayer(te3File, collisionLayer)
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

// Creates a map layer that doesn't render geometry.
func NewExtraMapLayer(te3File *te3.TE3File, collisionLayer collision.Mask) MapLayer {
	gridShape := collision.NewGrid(te3File.Tiles.Width, te3File.Tiles.Height, te3File.Tiles.Length, te3.GridSpacing)

	return MapLayer{
		Name:      te3File.FilePath(),
		Layer:     collisionLayer,
		GridShape: gridShape,
	}
}

func (gm *MapLayer) Update(deltaTime float32) {
	for i := range gm.tileAnims {
		gm.tileAnims[i].Update(deltaTime)
	}
}

func (gm *MapLayer) Render(context *render.Context) {
	for i := range gm.groupRenderers {
		gm.groupRenderers[i].Render(nil, &gm.tileAnims[i], context)
	}
}
