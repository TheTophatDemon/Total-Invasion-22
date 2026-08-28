package screens

import (
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

var MonthNames = [...]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}

type SavePreview struct {
	ui.Element
	SaveData *game.MapChangeSignal
	LabelKey string
}

func (sp *SavePreview) Init(labelKey string) {
	sp.Element = ui.NewBox(ui.Transform{
		Depth: 1,
	}, cache.GetTexture("assets/textures/ui/menu_background.png"))
	sp.BgColor = maybe.Some(color.White.WithAlpha(0.75))
	sp.SetTextWith("", ui.DefaultTextConfig().SetPadding(mgl32.Vec2{16.0, 16.0}))
	sp.LabelKey = labelKey
}

func (sp *SavePreview) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if sp == nil || sp.SaveData == nil {
		return
	}
	var summary strings.Builder
	if len(sp.LabelKey) > 0 {
		summary.WriteString(settings.Localize(sp.LabelKey))
		summary.WriteRune('\n')
		summary.WriteRune('\n')
	}
	if sp.SaveData.IsNil() {
		summary.WriteString(settings.Localize("saveFileEmpty"))
	} else {
		formattedTime := sp.SaveData.Timestamp.Format(settings.Localize("saveTimeFormat"))
		for _, month := range MonthNames {
			formattedTime = strings.Replace(formattedTime, month, settings.Localize("month"+month), 1)
		}
		summary.WriteString(formattedTime)
		summary.WriteRune('\n')
		summary.WriteString(strings.ToUpper(strings.TrimSuffix(sp.SaveData.MapTitleKey, "Title")))
		summary.WriteString(" - ")
		summary.WriteString(settings.Localize(sp.SaveData.MapTitleKey))
		summary.WriteRune('\n')
		if sp.SaveData.AfterCheckpoint {
			summary.WriteString(settings.Localize("afterCheckpoint"))
			summary.WriteRune('\n')
		}
	}
	sp.SetText(summary.String())
	queue.Add(&sp.Element)
}
