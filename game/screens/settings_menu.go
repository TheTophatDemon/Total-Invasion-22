package screens

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type (
	Resolution   [2]uint16
	OnOff        bool
	SettingsMenu struct {
		Menu
		chooserLanguage   Chooser[settings.Locale]
		chooserScreenSize Chooser[Resolution]
		chooserFullscreen Chooser[OnOff]
		chooserVsync      Chooser[OnOff]
		sliderSfxVolume   Slider
		sliderMusicVolume Slider
		sliderFOV         Slider
		sliderSensitivity Slider
		sliderTextShadow  Slider
		chooserChickens   Chooser[OnOff]
		bindingItem       MenuItem
		reinitNextFrame   bool
	}
)

func (res Resolution) String() string {
	return fmt.Sprintf("%vx%v", res[0], res[1])
}

func (onOff OnOff) String() string {
	if onOff {
		return settings.Localize("on")
	}
	return settings.Localize("off")
}

func (settingsMenu *SettingsMenu) Init(app engine.Observer, parent ui.Screen) *SettingsMenu {
	*settingsMenu = SettingsMenu{}

	settingsMenu.chooserLanguage.Init("language", []settings.Locale{settings.LocaleEnglish, settings.LocaleRussian}, settings.Current.Locale)

	sizeChoices := []Resolution{
		{640, 480},
		{800, 480},
		{800, 600},
		{1280, 720},
		{1920, 1080},
	}
	settingsMenu.chooserScreenSize.Init("screenResolution", sizeChoices, Resolution{settings.Current.WindowWidth, settings.Current.WindowHeight})

	onOffChoices := []OnOff{OnOff(false), OnOff(true)}
	settingsMenu.chooserFullscreen.Init("fullscreen", onOffChoices, OnOff(settings.Current.Fullscreen))
	settingsMenu.chooserVsync.Init("vsync", onOffChoices, OnOff(settings.Current.Vsync))

	settingsMenu.sliderSfxVolume.Init("sfxVolume", 0, 10, 1, settings.Current.SfxVolume*10.0)
	settingsMenu.sliderMusicVolume.Init("musVolume", 0, 10, 1, settings.Current.MusicVolume*10.0)
	settingsMenu.sliderFOV.Init("fov", 45, 120, 5, settings.Current.Fov)
	settingsMenu.sliderSensitivity.Init("mouseSensitivity", 1, 10, 1, settings.Current.MouseSensitivity)
	settingsMenu.sliderTextShadow.Init("textShadow", 0, 10, 1, settings.Current.TextShadowTransparency*10.0)

	settingsMenu.chooserChickens.Init("harmChickens", onOffChoices, OnOff(settings.Current.ChickenHarm))

	settingsMenu.bindingItem.Init("bindings", func(MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(BindingsMenu).Init(app, settingsMenu),
		})
	})

	settingsMenu.Menu.Init(
		app,
		[]MenuWidget{
			new(ReturnItem).Init(),
			&settingsMenu.chooserLanguage,
			&settingsMenu.chooserScreenSize,
			&settingsMenu.chooserFullscreen,
			&settingsMenu.chooserVsync,
			&settingsMenu.sliderSfxVolume,
			&settingsMenu.sliderMusicVolume,
			&settingsMenu.sliderFOV,
			&settingsMenu.sliderSensitivity,
			&settingsMenu.sliderTextShadow,
			&settingsMenu.chooserChickens,
			&settingsMenu.bindingItem,
		},
		parent,
	)
	return settingsMenu
}

func (menu *SettingsMenu) Layout(queue *ui.RenderQueue, deltaTime float32) {
	if menu.reinitNextFrame {
		menu.reinitNextFrame = false
		menu.Init(menu.app, menu.parent)
	}

	menu.Menu.Layout(queue, deltaTime)
	tdaudio.SetSfxVolume(float32(menu.sliderSfxVolume.FractionValue()))
	tdaudio.SetMusicVolume(float32(menu.sliderMusicVolume.FractionValue()))

	newTrans := menu.sliderTextShadow.FractionValue()
	settings.Current.TextShadowTransparency = newTrans
	ui.SetTextShadowColor(color.Black.WithAlpha(float32(newTrans)))

	if menu.chooserLanguage.Choice() != settings.Current.Locale {
		settings.Current.Locale = menu.chooserLanguage.Choice()
		menu.reinitNextFrame = true
	}
}

func (menu *SettingsMenu) Enter() {
}

func (menu *SettingsMenu) Exit() {
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
