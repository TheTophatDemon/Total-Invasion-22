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
		returnItem        MenuItem
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
		keybindItem       MenuItem
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
	settingsMenu.returnItem.Init("return", func(MenuInputType) { settingsMenu.returnToParent() })

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

	settingsMenu.keybindItem.Init("keybind", func(MenuInputType) {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(KeybindMenu).Init(app, settingsMenu),
		})
	})

	settingsMenu.Menu.Init(
		app,
		[]MenuEvents{
			&settingsMenu.returnItem,
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
			&settingsMenu.keybindItem,
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

func (menu *SettingsMenu) Exit() {
	originalSettings := settings.Current
	resetMenu := false

	if newLang := menu.chooserLanguage.Choice(); newLang != settings.Current.Locale {
		settings.Current.Locale = newLang
		resetMenu = true
	}

	newSize := menu.chooserScreenSize.Choice()
	if nowFullscreen := bool(menu.chooserFullscreen.Choice()); nowFullscreen != engine.IsFullscreen() {
		settings.Current.Fullscreen = nowFullscreen
		engine.SetFullscreen(nowFullscreen)
		resetMenu = true
	}
	if newSize[0] != settings.Current.WindowWidth || newSize[1] != settings.Current.WindowHeight {
		settings.Current.WindowWidth = newSize[0]
		settings.Current.WindowHeight = newSize[1]
		engine.SetScreenSize(int(settings.Current.WindowWidth), int(settings.Current.WindowHeight))
		resetMenu = true
	}
	if newVsync := bool(menu.chooserVsync.Choice()); settings.Current.Vsync != newVsync {
		settings.Current.Vsync = newVsync
		engine.SetVsync(newVsync)
	}

	if newVolume := menu.sliderSfxVolume.FractionValue(); !mgl32.FloatEqual(newVolume, settings.Current.SfxVolume) {
		settings.Current.SfxVolume = newVolume
	}
	if newVolume := menu.sliderMusicVolume.FractionValue(); !mgl32.FloatEqual(newVolume, settings.Current.MusicVolume) {
		settings.Current.MusicVolume = newVolume
	}

	if newFov := float32(menu.sliderFOV.IntValue()); !mgl32.FloatEqual(newFov, settings.Current.Fov) {
		settings.Current.Fov = newFov
	}

	if newSens := menu.sliderSensitivity.IntValue(); newSens != settings.Current.MouseSensitivity {
		settings.Current.MouseSensitivity = newSens

		// scaledSens := settings.ScaledMouseSensitivity(newSens)
		// input.ClearBinding(settings.ActionLookHorz)
		// input.BindActionMouseMove(settings.ActionLookHorz, input.MouseAxisX, scaledSens)
		// input.ClearBinding(settings.ActionLookVert)
		// input.BindActionMouseMove(settings.ActionLookVert, input.MouseAxisY, scaledSens)
		// TODO: How do we make this work again?
	}

	if chickenHarm := bool(menu.chooserChickens.Choice()); chickenHarm != settings.Current.ChickenHarm {
		settings.Current.ChickenHarm = chickenHarm
	}

	if resetMenu {
		// Reset the title menu so things are positioned correctly
		if titleMenu, ok := menu.parent.(*TitleMenu); ok {
			titleMenu.Init(titleMenu.app, titleMenu.inGame)
		}
	}
	if originalSettings != settings.Current {
		settings.Save()
	}
}
