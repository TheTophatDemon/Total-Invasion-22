package world

import (
	"strconv"

	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type Camera struct {
	comps.Camera
	id                  scene.Id[*Camera]
	linkNumber          int
	waitTimer, waitTime float32
}

var _ Linkable = (*Camera)(nil)

func SpawnCameraFromTE3(ent te3.Ent) (id scene.Id[*Camera], camera *Camera, err error) {
	id, camera, err = SpawnCamera(comps.TransformFromTE3Ent(ent, false, false))
	if err != nil {
		return
	}
	camera.id = id
	if linkStr, ok := ent.Properties["link"]; ok {
		var linkNo int64
		linkNo, err = strconv.ParseInt(linkStr, 10, 32)
		if err != nil {
			return
		}
		camera.linkNumber = int(linkNo)
	}
	if waitStr, ok := ent.Properties["wait"]; ok {
		if waitStr != "inf" && waitStr != "infinity" && waitStr != "-1" {
			var waitTime float64
			waitTime, err = strconv.ParseFloat(waitStr, 32)
			if err != nil {
				return
			}
			camera.waitTime = float32(waitTime)
		}
	} else {
		camera.waitTime = 2.0
	}
	return
}

func SpawnCamera(transform comps.Transform) (id scene.Id[*Camera], camera *Camera, err error) {
	id, camera, err = gWorld.Cameras.New()
	if err != nil {
		return
	}

	camera.Camera = comps.NewCamera(transform)

	return
}

func (camera *Camera) Update(deltaTime float32) {
	if camera.waitTime > 0.0 && gWorld.CurrentCamera.Equals(camera.id.Handle) {
		camera.waitTimer += deltaTime
		if camera.waitTimer > camera.waitTime {
			camera.waitTimer = 0.0
			gWorld.ResetToPlayerCamera()
		}
	}
}

func (camera *Camera) LinkNumber() int {
	return camera.linkNumber
}

func (camera *Camera) OnLinkActivate(source Linkable) {
	gWorld.CurrentCamera = camera.id
}

func (camera *Camera) OnLinkDeactivate(source Linkable) {
	if gWorld.CurrentCamera.Equals(camera.id.Handle) {
		gWorld.ResetToPlayerCamera()
	}
}

func (camera *Camera) Handle() scene.Handle {
	return camera.id.Handle
}
