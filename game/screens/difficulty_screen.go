package screens

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	DifficultyMenu struct {
		Menu
		instruction1, instruction2 ui.Element
	}
)

func (dm *DifficultyMenu) Init(app engine.Observer, parent ui.Screen) *DifficultyMenu {
	*dm = DifficultyMenu{}

	menuItems := make([]MenuWidget, 0, len(settings.Difficulties))

	for i := range settings.Difficulties {
		menuItems = append(menuItems, new(MenuItem).Init(
			fmt.Sprintf("difficulty%d", i),
			func(m *Menu, mw MenuWidget, mit MenuInputType) {
				settings.Current.DifficultyIndex = i
				app.ProcessSignal(game.MapChangeSignal{
					MapPath:       "assets/maps/e1m1.te3",
					SaveAfterLoad: true,
				})
			},
		))
	}

	dm.instruction1 = ui.NewText(ui.Transform{
		Depth:  20,
		Origin: ui.Ratios{0.5, 0},
	}, settings.Localize("chooseDifficulty"), ui.DefaultTextConfig())
	dm.instruction1.FitText()
	dm.instruction2 = ui.NewText(ui.Transform{
		Depth:  20,
		Origin: ui.Ratios{0.5, 1},
	}, settings.Localize("changeDifficulty"), ui.DefaultTextConfig())
	dm.instruction2.FitText()
	dm.Menu.Init(app, menuItems, parent)
	dm.Menu.background.SetWidth(5.0 * settings.UIWidth() / 8.0)
	return dm
}

func (dm *DifficultyMenu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	dm.Menu.Layout(queue, deltaTime)
	bounds := dm.Bounds()
	dm.instruction1.SetPosition(mgl32.Vec2{bounds.X + bounds.Width/2, bounds.Y + 8.0})
	dm.instruction2.SetPosition(mgl32.Vec2{bounds.X + bounds.Width/2, bounds.Y + bounds.Height - 8.0})
	queue.Add(&dm.instruction1)
	queue.Add(&dm.instruction2)
}
