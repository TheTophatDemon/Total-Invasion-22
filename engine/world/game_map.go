package world

import (
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

type Map struct {
	tiles          te3.Tiles
	mesh           *assets.Mesh
	tileAnims      []comps.AnimationPlayer
	groupRenderers []comps.MeshRender
}

func NewMap(te3File *te3.TE3File) (*Map, error) {
	mesh, err := te3File.BuildMesh()
	if err != nil {
		return nil, err
	}
	assets.TakeMesh(te3File.FilePath(), mesh)

	gameMap := &Map{
		tiles:          te3File.Tiles,
		mesh:           mesh,
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
