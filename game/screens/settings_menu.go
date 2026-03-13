package screens

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
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

	settingsMenu.sliderSfxVolume.Init("sfxVolume", 0, 10, 1, int(settings.Current.SfxVolume*10.0))
	settingsMenu.sliderMusicVolume.Init("musVolume", 0, 10, 1, int(settings.Current.MusicVolume*10.0))
	settingsMenu.sliderFOV.Init("fov", 45, 120, 5, int(settings.Current.Fov))
	settingsMenu.sliderSensitivity.Init("mouseSensitivity", 1, 10, 1, min(10, max(1, settings.Current.MouseSensitivity)))
	settingsMenu.sliderTextShadow.Init("textShadow", 0, 10, 1, int(settings.Current.TextShadowTransparency*10.0))

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
	menu.Menu.Layout(queue, deltaTime)
	tdaudio.SetSfxVolume(menu.sliderSfxVolume.FractionValue())
	tdaudio.SetMusicVolume(menu.sliderMusicVolume.FractionValue())

	if newTrans := float32(menu.sliderTextShadow.IntValue()) / 10.0; !mgl32.FloatEqual(newTrans, settings.Current.TextShadowTransparency) {
		settings.Current.TextShadowTransparency = newTrans
		ui.SetTextShadowColor(color.Black.WithAlpha(newTrans))
	}
}

func (menu *SettingsMenu) Enter() {
}

func (menu *SettingsMenu) Exit() {
	settings.Current.Locale = menu.chooserLanguage.Choice()

	newSize := menu.chooserScreenSize.Choice()

	settings.Current.Fullscreen = bool(menu.chooserFullscreen.Choice())
	engine.SetFullscreen(settings.Current.Fullscreen)

	settings.Current.WindowWidth = newSize[0]
	settings.Current.WindowHeight = newSize[1]
	engine.SetScreenSize(int(settings.Current.WindowWidth), int(settings.Current.WindowHeight))

	settings.Current.Vsync = bool(menu.chooserVsync.Choice())
	engine.SetVsync(settings.Current.Vsync)

	settings.Current.SfxVolume = menu.sliderSfxVolume.FractionValue()
	settings.Current.MusicVolume = menu.sliderMusicVolume.FractionValue()
	settings.Current.Fov = float32(menu.sliderFOV.IntValue())

	settings.Current.MouseSensitivity = menu.sliderSensitivity.IntValue()
	settings.Current.ChickenHarm = bool(menu.chooserChickens.Choice())

	// Reset the title menu so things are positioned correctly
	if titleMenu, ok := menu.parent.(*TitleMenu); ok {
		titleMenu.Init(titleMenu.app, titleMenu.inGame)
	}
	settings.Save()
}
