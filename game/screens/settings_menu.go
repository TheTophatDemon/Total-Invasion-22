package screens

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	Resolution   [2]uint16
	Pixelization int
	OnOff        bool
	SettingsMenu struct {
		Menu
		chooserDifficulty              Chooser[settings.Difficulty]
		chooserLanguage                Chooser[settings.Locale]
		chooserScreenSize              Chooser[Resolution]
		chooserPixelization            Chooser[Pixelization]
		chooserFullscreen              Chooser[OnOff]
		chooserVsync                   Chooser[OnOff]
		sliderSfxVolume                Slider
		sliderMusicVolume              Slider
		sliderFOV                      Slider
		sliderSensitivity              Slider
		sliderTextShadow               Slider
		chooserChickens                Chooser[OnOff]
		bindingItem                    MenuItem
		reinitNextFrame                bool
		wraith, difficultyWarningLabel ui.Element
	}
)

func (res Resolution) String() string {
	return fmt.Sprintf("%vx%v", res[0], res[1])
}

func (pix Pixelization) String() string {
	switch pix {
	case 1:
		return settings.Localize("noPixelization")
	case 2:
		return settings.Localize("somePixelization")
	case 4:
		return settings.Localize("hardPixelization")
	case 8:
		return settings.Localize("maxPixelization")
	default:
		return "INVALID"
	}
}

func (onOff OnOff) String() string {
	if onOff {
		return settings.Localize("on")
	}
	return settings.Localize("off")
}

func (settingsMenu *SettingsMenu) Init(app engine.Observer, parent ui.Screen) *SettingsMenu {

	*settingsMenu = SettingsMenu{}

	inGame := parent.(*TitleMenu).inGame
	if inGame {
		settingsMenu.chooserDifficulty.Init("difficulty", settings.Difficulties[:], settings.CurrDifficulty())
		settingsMenu.difficultyWarningLabel = ui.NewText(
			ui.Transform{
				Depth: 20,
			},
			settings.Localize("difficultyWarning"),
			ui.DefaultTextConfig().SetAlign(ui.TextAlignCenterH).SetColor(color.Red),
		)
		settingsMenu.difficultyWarningLabel.FitText()
	}

	settingsMenu.chooserLanguage.Init("language", []settings.Locale{settings.LocaleEnglish, settings.LocaleRussian}, settings.Current.Locale)

	sizeChoices := []Resolution{
		{640, 480},
		{800, 480},
		{800, 600},
		{1280, 720},
		{1920, 1080},
	}
	settingsMenu.chooserScreenSize.Init("screenResolution", sizeChoices, Resolution{settings.Current.WindowWidth, settings.Current.WindowHeight})

	pixelizationChoices := []Pixelization{1, 2, 4, 8}
	settingsMenu.chooserPixelization.Init("pixelization", pixelizationChoices, Pixelization(settings.Current.Pixelization))

	onOffChoices := []OnOff{OnOff(false), OnOff(true)}
	settingsMenu.chooserFullscreen.Init("fullscreen", onOffChoices, OnOff(settings.Current.Fullscreen))
	settingsMenu.chooserVsync.Init("vsync", onOffChoices, OnOff(settings.Current.Vsync))

	settingsMenu.sliderSfxVolume.Init("sfxVolume", 0, 10, 1, settings.Current.SfxVolume*10.0)
	settingsMenu.sliderMusicVolume.Init("musVolume", 0, 10, 1, settings.Current.MusicVolume*10.0)
	settingsMenu.sliderFOV.Init("fov", 45, 120, 5, settings.Current.Fov)
	settingsMenu.sliderSensitivity.Init("mouseSensitivity", 1, 10, 1, settings.Current.MouseSensitivity)
	settingsMenu.sliderTextShadow.Init("textShadow", 0, 10, 1, settings.Current.TextShadowTransparency*10.0)

	settingsMenu.chooserChickens.Init("harmChickens", onOffChoices, OnOff(settings.Current.ChickenHarm))

	settingsMenu.bindingItem.Init("bindings", func(menu *Menu, item MenuWidget, mit MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(BindingsMenu).Init(app, settingsMenu),
		})
	})

	widgets := make([]MenuWidget, 0, 14)
	widgets = append(widgets, new(ReturnItem).Init())
	if inGame {
		widgets = append(widgets, &settingsMenu.chooserDifficulty)
	}
	widgets = append(widgets,
		&settingsMenu.chooserLanguage,
		&settingsMenu.chooserScreenSize,
		&settingsMenu.chooserPixelization,
		&settingsMenu.chooserFullscreen,
		&settingsMenu.chooserVsync,
		&settingsMenu.sliderSfxVolume,
		&settingsMenu.sliderMusicVolume,
		&settingsMenu.sliderFOV,
		&settingsMenu.sliderSensitivity,
		&settingsMenu.sliderTextShadow,
		&settingsMenu.chooserChickens,
		&settingsMenu.bindingItem,
	)

	settingsMenu.Menu.Init(app, widgets, parent)

	wraithTex := cache.GetTexture("assets/textures/sprites/wraith.png")
	if headbangAnim, ok := wraithTex.GetAnimation("headbang"); ok {
		settingsMenu.wraith = ui.NewBox(ui.Transform{
			Position: settingsMenu.position.Add(mgl32.Vec2{}),
			Size:     mgl32.Vec2{64.0, 64.0},
			Origin:   ui.Ratios{1.0, 0.0},
			Depth:    20,
		}, wraithTex)
		settingsMenu.wraith.AnimPlayer.PlayNewAnim(headbangAnim)
	}

	return settingsMenu
}

func (menu *SettingsMenu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if menu.reinitNextFrame {
		menu.reinitNextFrame = false
		menu.Init(menu.app, menu.parent)
	}

	menu.Menu.Layout(queue, deltaTime)
	bounds := menu.Bounds()
	menu.wraith.SetPosition(mgl32.Vec2{bounds.X + bounds.Width, bounds.Y})
	menu.wraith.AnimPlayer.Update(deltaTime)
	queue.Add(&menu.wraith)

	tdaudio.SetSfxVolume(float32(menu.sliderSfxVolume.FractionValue()))
	tdaudio.SetMusicVolume(float32(menu.sliderMusicVolume.FractionValue()))

	newTrans := menu.sliderTextShadow.FractionValue()
	settings.Current.TextShadowTransparency = newTrans
	ui.SetTextShadowColor(color.Black.WithAlpha(float32(newTrans)))

	if menu.chooserLanguage.Choice() != settings.Current.Locale {
		settings.Current.Locale = menu.chooserLanguage.Choice()
		menu.reinitNextFrame = true
	}

	if len(menu.chooserDifficulty.choices) > 0 && menu.chooserDifficulty.Choice().Index != settings.Current.DifficultyIndex {
		menu.difficultyWarningLabel.SetPosition(bounds.PosVec().Add(mgl32.Vec2{48.0, 8.0}))
		menu.difficultyWarningLabel.SetWidth(bounds.Width - 96.0)
		menu.difficultyWarningLabel.SetHeight(64.0)
		queue.Add(&menu.difficultyWarningLabel)
	}
}

func (menu *SettingsMenu) Enter() {
}

func (menu *SettingsMenu) Exit() {
	if len(menu.chooserDifficulty.choices) > 0 {
		settings.Current.DifficultyIndex = menu.chooserDifficulty.Choice().Index
	}
	settings.Current.Locale = menu.chooserLanguage.Choice()
	settings.Current.Vsync = bool(menu.chooserVsync.Choice())
	settings.Current.SfxVolume = menu.sliderSfxVolume.FractionValue()
	settings.Current.MusicVolume = menu.sliderMusicVolume.FractionValue()
	settings.Current.Fov = menu.sliderFOV.FloatValue()
	settings.Current.MouseSensitivity = menu.sliderSensitivity.FloatValue()
	settings.Current.ChickenHarm = bool(menu.chooserChickens.Choice())

	settingsBeforeSizeChange := settings.Current
	settings.Current.Fullscreen = bool(menu.chooserFullscreen.Choice())
	newSize := menu.chooserScreenSize.Choice()
	settings.Current.WindowWidth = newSize[0]
	settings.Current.WindowHeight = newSize[1]
	settings.Current.Pixelization = int(menu.chooserPixelization.Choice())
	settings.Current.Apply()

	gotLarger := settingsBeforeSizeChange.WindowWidth < settings.Current.WindowWidth ||
		settingsBeforeSizeChange.WindowHeight < settings.Current.WindowHeight ||
		(!settingsBeforeSizeChange.Fullscreen && settings.Current.Fullscreen)
	if gotLarger {
		menu.app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(ConfirmMenu).Init(menu.app, menu, menu.parent, settingsBeforeSizeChange),
		})
	}

	// Reset the title menu so things are positioned correctly
	if titleMenu, ok := menu.parent.(*TitleMenu); ok {
		titleMenu.Init(titleMenu.app, titleMenu.inGame)
	}
	settings.Save()
}
