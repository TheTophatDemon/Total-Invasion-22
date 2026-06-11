package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"

	"github.com/go-gl/mathgl/mgl32"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/engine/timer"
	"tophatdemon.com/total-invasion-ii/game"

	"tophatdemon.com/total-invasion-ii/game/screens"
	"tophatdemon.com/total-invasion-ii/game/settings"
	"tophatdemon.com/total-invasion-ii/game/world"
)

type App struct {
	world                 *world.World
	renderQueue           ui.RenderQueue
	screen                ui.Screen
	loadingMap            game.MapChangeSignal
	loadingTimer          timer.Timer // Waits until the loading screen renders to load the map
	titleScreenBackground ui.Element
	signalQueue           []any
}

func (app *App) Update(deltaTime float32) {
	if settings.ActionLevelSelect.JustPressed() {
		app.ProcessSignal(game.ChangeScreenSignal{
			Screen: new(screens.LevelSelectMenu).Init(app),
		})
	}

	if app.screen == nil && app.world == nil && app.loadingTimer.Update(deltaTime) {
		app.LoadGame(app.loadingMap)
	}

	switch true {
	case app.screen != nil:
		app.renderQueue.Clear()
		// Render title screen graphic
		app.titleScreenBackground = ui.NewBox(
			ui.Transform{},
			cache.GetTexture("assets/textures/ui/title_screen.png"),
		)
		app.titleScreenBackground.FitHeight(float32(settings.Current.WindowHeight))
		app.renderQueue.Add(&app.titleScreenBackground)
		app.screen.Layout(&app.renderQueue, deltaTime)
	case app.world != nil:
		app.world.Update(deltaTime)
		if settings.Current.ActionMenuCancel.JustPressed() {
			input.UntrapMouse()
			app.screen = new(screens.TitleMenu).Init(app, true)
		}
	}

	// Process signals
	for len(app.signalQueue) > 0 {
		lastIdx := len(app.signalQueue) - 1
		signal := app.signalQueue[lastIdx]
		app.signalQueue = app.signalQueue[:lastIdx]
		app.executeSignal(signal)
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
		renderContext.Enable2D()
		app.renderQueue.Render(&renderContext, false)
	case app.world != nil:
		app.world.Render()
	default:
		// Draw the loading screen
		renderContext := render.Context{
			View:       mgl32.Ident4(),
			Projection: mgl32.Ortho(0.0, float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight), 0.0, -50.0, 50.0),
		}
		renderContext.Enable2D()
		loadingScreenTex := cache.GetTexture(fmt.Sprintf("assets/textures/ui/loading_screen_%v.png", string(settings.Current.Locale)))
		box := ui.NewBox(
			ui.Transform{
				Anchor: ui.Ratios{0.5, 0.0},
				Origin: ui.Ratios{0.5, 0.0},
				Depth:  10.0,
			},
			loadingScreenTex,
		)
		box.FitHeight(float32(settings.Current.WindowHeight))
		box.Render(&renderContext)
	}
}

func (app *App) ProcessSignal(signal any) {
	app.signalQueue = append(app.signalQueue, signal)
}

func (app *App) executeSignal(signal any) {
	switch msg := signal.(type) {
	case game.ResumeGameSignal:
		if app.world != nil {
			app.screen = nil
			input.TrapMouse()
			app.world.ProcessSignal(signal)
		}
	case game.MapChangeSignal:
		if app.world != nil {
			app.world.TearDown()
		}
		app.world = nil
		app.screen = nil
		app.loadingMap = msg
		app.loadingTimer = timer.Timer{
			Interval: 0.5, // Make the player wait at least a bit to look at my amazing artwork ;-)
			MaxTicks: 1,
		}
	case game.ChangeScreenSignal:
		if app.screen != nil {
			app.screen.Exit()
		}
		if msg.Screen != nil {
			app.screen = msg.Screen
			msg.Screen.Enter()
			input.UntrapMouse()
		} else if app.world != nil {
			// No parent screen = resuming game
			input.TrapMouse()
			app.screen = nil
		}
	}
}

func (app *App) LoadGame(sig game.MapChangeSignal) {
	log.Println("Loading game at map ", sig.NextMapPath)

	cache.Reset()

	world, err := world.NewWorld(app, sig)
	if err != nil {
		panic(err)
	}

	input.TrapMouse()
	app.world = world

	runtime.GC()
}

func main() {
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

	engine.Init(
		int(settings.Current.WindowWidth),
		int(settings.Current.WindowHeight),
		"Total Invasion 22",
		slices.Contains(os.Args[1:], "debug"),
		settings.Current.Fullscreen,
	)
	defer engine.DeInit()
	settings.Current.Apply()

	// The first sound loaded is used as the error sound
	cache.GetSfx("assets/sounds/error.wav")
	cache.PreloadSfx("assets/sounds")

	cache.DefaultFont, _ = cache.GetFont("assets/textures/ui/font.fnt")

	app := &App{}
	mapName := settings.Current.Debug.StartMap
	if len(mapName) == 0 {
		app.screen = screens.NewIntroScreen(app)
		tdaudio.QueueSong("assets/music/back_in_vasion.ogg", true, 0)
	} else {
		app.executeSignal(game.MapChangeSignal{
			NextMapPath: mapName,
		})
	}
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
