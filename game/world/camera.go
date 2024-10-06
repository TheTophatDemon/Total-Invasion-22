package world

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Camera struct {
	comps.Camera
	linkNumber int
}

var _ Linkable = (*Camera)(nil)

func SpawnCameraFromTE3(world *World, ent te3.Ent) (id scene.Id[*Camera], camera *Camera, err error) {
	id, camera, err = SpawnCamera(world, comps.TransformFromTE3Ent(ent, false, false))
	if err != nil {
		return
	}
	camera.linkNumber, _ = ent.IntProperty("link")
	return
}

func SpawnCamera(world *World, transform comps.Transform) (id scene.Id[*Camera], camera *Camera, err error) {
	id, camera, err = world.Cameras.New()
	if err != nil {
		return
	}

	camera.Camera = comps.NewCamera(
		settings.Current.Fov, settings.Current.WindowAspectRatio(), 0.1, 1000.0, transform,
	)

	return
}

func (camera *Camera) LinkNumber() int {
	return camera.linkNumber
}
