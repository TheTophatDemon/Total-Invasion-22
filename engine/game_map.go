package engine

import (
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/ecomps"
	"tophatdemon.com/total-invasion-ii/engine/ecs"
)

// Creates the entities in the scene required to render a map from the given te3 file.
func SpawnGameMap(scene *ecs.Scene, te3File *assets.TE3File) ([]ecs.Entity, error) {
	mesh, err := te3File.BuildMesh()
	if err != nil {
		return nil, err
	}
	assets.TakeMesh(te3File.FilePath(), mesh)

	//Create tile animation players
	mapEnts := make([]ecs.Entity, mesh.GetGroupCount())
	for g, groupName := range mesh.GetGroupNames() {
		mapEnts[g], err = scene.AddEntity()
		if err != nil {
			return nil, err
		}

		tex := assets.GetTexture(groupName)
		if tex.AnimationCount() > 0 {
			// Add animations if applicable
			ecomps.AddAnimationPlayer(mapEnts[g], tex.GetAnimation(0), true)
		}

		// Add mesh component
		ecomps.AddMeshRenderGroup(mapEnts[g], mesh, assets.MapShader, tex, groupName)
	}

	return mapEnts, nil
}
