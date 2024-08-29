package world

import (
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
)

type Effect struct {
	Transform comps.Transform
	Lifetime  float32

	world     *World
	id        scene.Id[*Effect]
	particles comps.ParticleRender
}

func SpawnEffect(world *World, storage *scene.Storage[Effect], transform comps.Transform, lifetime float32, particles comps.ParticleRender) (id scene.Id[*Effect], fx *Effect, err error) {
	id, fx, err = storage.New()
	if err != nil {
		return
	}

	*fx = Effect{
		Transform: transform,
		Lifetime:  lifetime,
		world:     world,
		id:        id,
		particles: particles,
	}

	fx.particles.Init()
	return
}

func (fx *Effect) Finalize() {
	fx.particles.Finalize()
}

func (fx *Effect) Update(deltaTime float32) {
	fx.Lifetime -= deltaTime
	fx.particles.Update(deltaTime, &fx.Transform)
	if fx.Lifetime < 0.0 {
		fx.world.QueueRemoval(fx.id.Handle)
	}
}

func (fx *Effect) Render(context *render.Context) {
	fx.particles.Render(&fx.Transform, context)
}
