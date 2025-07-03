package hud

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

type DebugStats struct {
	fpsCounter, drawCounters ui.Text
}

func (stats *DebugStats) init() {
	stats.fpsCounter = ui.Text{
		Transform: ui.Transform{
			Dest: math2.Rect{X: 4.0, Y: 20.0, Width: 160.0, Height: 32.0},
		},
	}

	stats.drawCounters = ui.Text{
		Transform: ui.Transform{
			Dest: math2.Rect{X: 4.0, Y: 56.0, Width: 480.0, Height: 128.0},
		},
		Color: color.Blue,
	}
}

func (stats *DebugStats) Layout(queue *ui.RenderQueue) {
	stats.fpsCounter.SetText(fmt.Sprintf("FPS: %v", engine.FPS()))
	queue.Add(&stats.fpsCounter)

	queue.Add(&stats.drawCounters)
}

func (stats *DebugStats) UpdateCounters(renderContext *render.Context, avgCollisionTime int64) {
	stats.drawCounters.SetText(
		fmt.Sprintf("Sprites drawn: %v\nWalls drawn: %v\nParticles drawn: %v\nAvg. Collision MS: %v",
			renderContext.DrawnSpriteCount,
			renderContext.DrawnWallCount,
			renderContext.DrawnParticlesCount,
			avgCollisionTime))
}
