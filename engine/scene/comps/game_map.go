package comps

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
)

type Map struct {
	GridShape      collision.Grid
	name           string
	body           Body
	tiles          te3.Tiles
	mesh           *geom.Mesh
	triMap         te3.TriMap        // Maps a flattened tile index to its indices in the mesh's triangles array.
	tileAnims      []AnimationPlayer // Animates each texture group of tiles
	groupRenderers []MeshRender      // Renders each texture group of tiles
}

var _ HasBody = (*Map)(nil)

func NewMap(te3File *te3.TE3File, collisionLayer collision.Mask) (Map, error) {
	mesh, triMap, err := te3File.BuildMesh()
	if err != nil {
		return Map{}, err
	}
	cache.TakeMesh(te3File.FilePath(), mesh)

	var gridShape collision.Grid = collision.NewGrid(te3File.Tiles.Width, te3File.Tiles.Height, te3File.Tiles.Length, te3File.Tiles.GridSpacing())

	gameMap := Map{
		name: te3File.FilePath(),
		body: Body{
			Shape: gridShape,
			Layer: collisionLayer,
		},
		GridShape:      gridShape,
		tiles:          te3File.Tiles,
		mesh:           mesh,
		triMap:         triMap,
		tileAnims:      make([]AnimationPlayer, mesh.GroupCount()),
		groupRenderers: make([]MeshRender, mesh.GroupCount()),
	}

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

func (gameMap *Map) Name() string {
	return gameMap.name
}

func (gameMap *Map) Body() *Body {
	return &gameMap.body
}

func (gameMap *Map) TearDown() {
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
