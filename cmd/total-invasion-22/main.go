package main

import (
	"log"
	"os"
	"runtime"
	"slices"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/engine/timer"
	"tophatdemon.com/total-invasion-ii/game"

	"tophatdemon.com/total-invasion-ii/game/screens"
	"tophatdemon.com/total-invasion-ii/game/settings"
	"tophatdemon.com/total-invasion-ii/game/world"
)

type App struct {
	world        *world.World
	renderQueue  ui.RenderQueue
	screen       *screens.Screen
	loadingMap   game.MapChangeSignal
	loadingTimer timer.Timer // Waits until the loading screen renders to load the map
}

func (app *App) Update(deltaTime float32) {
	// Update audio volume based on settings.
	tdaudio.SetSfxVolume(settings.Current.SfxVolume)
	tdaudio.SetMusicVolume(settings.Current.MusicVolume)

	switch true {
	case app.screen != nil:
		app.screen.Layout(&app.renderQueue, deltaTime)
	case app.world != nil:
		app.world.Update(deltaTime)
	case app.loadingTimer.Update(deltaTime):
		app.LoadGame(app.loadingMap)
	}
}

func (app *App) Render() {
	switch true {
	case app.screen != nil:
		// Setup 2D render context
		renderContext := render.Context{
			View:       mgl32.Ident4(),
			Projection: mgl32.Ortho(0.0, float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight), 0.0, -50.0, 50.0),
		}
		app.renderQueue.Render(&renderContext)
	case app.world != nil:
		app.world.Render()
	default:
		// Draw the loading screen
		renderContext := render.Context{
			View:       mgl32.Ident4(),
			Projection: mgl32.Ortho(0.0, float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight), 0.0, -50.0, 50.0),
		}
		loadingScreenTex := cache.GetTexture("assets/textures/ui/loading_screen_" + settings.Current.Locale + ".png")
		gl.CullFace(gl.FRONT)
		box := ui.NewBox(
			ui.Transform{
				Size:  mgl32.Vec2{settings.UIWidth(), settings.UIHeight()},
				Depth: 10.0,
			},
			loadingScreenTex,
		)
		box.Render(&renderContext)
	}
}

func (app *App) ProcessSignal(signal any) {
	switch msg := signal.(type) {
	case game.MapChangeSignal:
		if app.world != nil {
			app.world.TearDown()
		}
		app.world = nil
		app.loadingMap = msg
		app.loadingTimer = timer.Timer{
			Interval: 0.5, // Make the player wait at least a bit to look at my amazing artwork ;-)
			MaxTicks: 1,
		}
	}
}

func (app *App) LoadGame(sig game.MapChangeSignal) {
	log.Println("Loading game at map ", sig.NextMapPath)

	cache.Reset()

	world, err := world.NewWorld(app, sig.NextMapPath, sig)
	if err != nil {
		panic(err)
	}

	input.TrapMouse()
	app.world = world

	runtime.GC()
}

func main() {
	var err error
	// cpuProfile, err := os.Create("cpuProfile.pprof")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer cpuProfile.Close()
	// if err := pprof.StartCPUProfile(cpuProfile); err != nil {
	// 	log.Fatal(err)
	// }
	// defer pprof.StopCPUProfile()

	settings.LoadOrInit()

	err = engine.Init(
		int(settings.Current.WindowWidth),
		int(settings.Current.WindowHeight),
		"Total Invasion 22",
		slices.Contains(os.Args[1:], "debug"),
	)
	defer engine.DeInit()
	if err != nil {
		panic(err)
	}

	// The first sound loaded is used as the error sound
	cache.GetSfx("assets/sounds/error.wav")
	cache.PreloadSfx("assets/sounds")

	input.BindActionKey(settings.ActionForward, glfw.KeyW)
	input.BindActionKey(settings.ActionBack, glfw.KeyS)
	input.BindActionKey(settings.ActionLeft, glfw.KeyA)
	input.BindActionKey(settings.ActionRight, glfw.KeyD)
	input.BindActionKey(settings.ActionSlow, glfw.KeyLeftShift)
	input.BindActionKey(settings.ActionTrapMouse, glfw.KeyEscape)
	input.BindActionKey(settings.ActionUse, glfw.KeyE)
	input.BindActionMouseMove(settings.ActionLookHorz, input.MouseAxisX, settings.Current.MouseSensitivity)
	input.BindActionMouseMove(settings.ActionLookVert, input.MouseAxisY, settings.Current.MouseSensitivity)
	input.BindActionMouseButton(settings.ActionFire, glfw.MouseButton1)
	input.BindActionKey(settings.ActionWeaponWheel, glfw.KeyQ)
	input.BindActionKey(settings.ActionSickle, glfw.Key1)
	input.BindActionKey(settings.ActionChicken, glfw.Key2)
	input.BindActionKey(settings.ActionGrenade, glfw.Key3)
	input.BindActionKey(settings.ActionParusu, glfw.Key4)
	input.BindActionKey(settings.ActionDblGrenade, glfw.Key5)
	input.BindActionKey(settings.ActionSign, glfw.Key6)
	input.BindActionKey(settings.ActionAirhorn, glfw.Key7)
	input.BindActionKey(settings.ActionDefenestrator, glfw.Key8)
	input.BindActionKey(settings.ActionCluckster, glfw.Key9)
	input.BindActionCharSequence(settings.ActionNoclip, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyC, glfw.KeyL, glfw.KeyI, glfw.KeyP})                               //TDCLIP
	input.BindActionCharSequence(settings.ActionGodMode, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyD, glfw.KeyQ, glfw.KeyD})                                         //TDDQD
	input.BindActionCharSequence(settings.ActionMarySue, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyM, glfw.KeyS, glfw.KeyM})                                         //TDMSM
	input.BindActionCharSequence(settings.ActionDie, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyU, glfw.KeyN, glfw.KeyA, glfw.KeyL, glfw.KeyI, glfw.KeyV, glfw.KeyE}) //TDUNALIVE
	input.BindActionCharSequence(settings.ActionKillEnemies, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyN, glfw.KeyU, glfw.KeyK, glfw.KeyE, glfw.KeyM})               //TDNUKEM
	input.BindActionCharSequence(settings.ActionCastBlessing, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyW, glfw.KeyO, glfw.KeyL, glfw.KeyO, glfw.KeyL, glfw.KeyO})   //TDWOLOLO
	input.BindActionCharSequence(settings.ActionLaunchEditor, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyJ, glfw.KeyO, glfw.KeyM, glfw.KeyT})                         //TDJOMT
	input.BindActionCharSequence(settings.ActionSpawnChicken, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyK, glfw.KeyF, glfw.KeyC})                                    //TDKFC

	input.BindActionKey(settings.ActionMenuUp, glfw.KeyUp)
	input.BindActionKey(settings.ActionMenuDown, glfw.KeyDown)
	input.BindActionKey(settings.ActionMenuConfirm, glfw.KeyEnter)
	// input.BindActionMouseButton(settings.ActionMenuConfirm, glfw.MouseButton1)
	input.BindActionKey(settings.ActionMenuCancel, glfw.KeyEscape)

	cache.DefaultFont, _ = cache.GetFont("assets/textures/ui/font.fnt")

	titleScreen := screens.Screen{}
	titleScreen.InitTitleScreen(mgl32.Vec2{32.0, 32.0}, false)

	app := &App{
		screen: &titleScreen,
	}
	// mapName := settings.Current.Debug.StartMap
	// if len(mapName) == 0 {
	// 	mapName = "assets/maps/e1m1-genocide-carnival.te3"
	// }
	// app.ProcessSignal(game.MapChangeSignal{
	// 	NextMapPath: mapName,
	// })
	engine.Run(app)

	// memProf, err := os.Create("memory_profile.pprof")
	// if err != nil {
	// 	log.Fatalf("could not create memory profile: %v", err)
	// }
	// defer memProf.Close()
	// runtime.GC()
	// if err := pprof.WriteHeapProfile(memProf); err != nil {
	// 	log.Fatal("could not write memory profile: ", err)
	// }
}
