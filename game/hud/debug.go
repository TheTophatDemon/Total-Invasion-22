package hud

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
)

type DebugStats struct {
	fpsCounter, drawCounters ui.Element
}

func (stats *DebugStats) init() {
	config := ui.TextConfig{
		Color: maybe.Some(color.Blue),
	}
	stats.fpsCounter = ui.NewText(
		ui.Transform{
			Position: mgl32.Vec2{4.0, 52.0},
			Size:     mgl32.Vec2{160.0, 32.0},
		},
		"",
		config,
	)

	stats.drawCounters = ui.NewText(
		ui.Transform{
			Position: mgl32.Vec2{4.0, 88.0},
			Size:     mgl32.Vec2{480.0, 128.0},
		},
		"",
		config,
	)
}

func (stats *DebugStats) Layout(queue *ui.RenderQueue) {
	stats.fpsCounter.SetText(fmt.Sprintf("FPS: %v", engine.FPS()))
	queue.Add(&stats.fpsCounter)

	queue.Add(&stats.drawCounters)
}

func (stats *DebugStats) UpdateCounters(renderContext *render.Context) {
	stats.drawCounters.SetText(
		fmt.Sprintf("Sprites drawn: %v\nWalls drawn: %v\nParticles drawn: %v",
			renderContext.DrawnSpriteCount,
			renderContext.DrawnWallCount,
			renderContext.DrawnParticlesCount))
}
