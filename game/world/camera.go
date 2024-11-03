package world

import (
	"strconv"

	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Camera struct {
	comps.Camera
	world               *World
	id                  scene.Id[*Camera]
	linkNumber          int
	waitTimer, waitTime float32
}

var _ Linkable = (*Camera)(nil)

func SpawnCameraFromTE3(world *World, ent te3.Ent) (id scene.Id[*Camera], camera *Camera, err error) {
	id, camera, err = SpawnCamera(world, comps.TransformFromTE3Ent(ent, false, false))
	if err != nil {
		return
	}
	camera.world = world
	camera.id = id
	if linkStr, ok := ent.Properties["link"]; ok {
		var linkNo int64
		linkNo, err = strconv.ParseInt(linkStr, 10, 32)
		if err != nil {
			return
		}
		camera.linkNumber = int(linkNo)
	}
	if waitStr, ok := ent.Properties["wait"]; ok && waitStr != "inf" && waitStr != "infinity" && waitStr != "-1" {
		var waitTime float64
		waitTime, err = strconv.ParseFloat(waitStr, 32)
		if err != nil {
			return
		}
		camera.waitTime = float32(waitTime)
	}
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

func (camera *Camera) Update(deltaTime float32) {
	if camera.waitTime > 0.0 && camera.world.CurrentCamera.Equals(camera.id.Handle) {
		camera.waitTimer += deltaTime
		if camera.waitTimer > camera.waitTime {
			camera.waitTimer = 0.0
			if player, isPlayer := camera.world.CurrentPlayer.Get(); isPlayer {
				camera.world.CurrentCamera = player.Camera
			}
		}
	}
}

func (camera *Camera) LinkNumber() int {
	return camera.linkNumber
}

func (camera *Camera) OnLinkActivate(source Linkable) {
	camera.world.CurrentCamera = camera.id
}

func (camera *Camera) Handle() scene.Handle {
	return camera.id.Handle
}
