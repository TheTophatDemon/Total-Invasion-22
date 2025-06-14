package main

import (
	"log"
	"os"
	"runtime"
	"slices"

	"github.com/go-gl/glfw/v3.3/glfw"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/input"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"

	"tophatdemon.com/total-invasion-ii/game/settings"
	"tophatdemon.com/total-invasion-ii/game/world"
)

type App struct {
	world *world.World
}

func (app *App) Update(deltaTime float32) {
	// Update audio volume based on settings.
	tdaudio.SetSfxVolume(settings.Current.SfxVolume)
	tdaudio.SetMusicVolume(settings.Current.MusicVolume)

	app.world.Update(deltaTime)
}

func (app *App) Render() {
	app.world.Render()
}

func (app *App) ProcessSignal(signal any) {
	switch msg := signal.(type) {
	case game.MapChangeSignal:
		if app.world != nil {
			app.world.TearDown()
		}
		app.LoadGame(msg.NextMapPath)
	}
}

func (app *App) LoadGame(mapPath string) {
	log.Println("Loading game at map ", mapPath)

	cache.Reset()
	cache.DefaultFont, _ = cache.GetFont(world.DEFAULT_FONT_PATH)

	debugMode := slices.Contains(os.Args[1:], "debug")
	world, err := world.NewWorld(app, mapPath, debugMode)
	if err != nil {
		panic(err)
	}

	if !debugMode {
		input.TrapMouse()
	}
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

	err = engine.Init(int(settings.Current.WindowWidth), int(settings.Current.WindowHeight), "Total Invasion 22")
	defer engine.DeInit()
	if err != nil {
		panic(err)
	}

	// Load error sound as first sound
	tdaudio.LoadSound("assets/sounds/error.wav", 1, false, 1.0)

	input.BindActionKey(settings.ACTION_FORWARD, glfw.KeyW)
	input.BindActionKey(settings.ACTION_BACK, glfw.KeyS)
	input.BindActionKey(settings.ACTION_LEFT, glfw.KeyA)
	input.BindActionKey(settings.ACTION_RIGHT, glfw.KeyD)
	input.BindActionKey(settings.ACTION_SLOW, glfw.KeyLeftShift)
	input.BindActionKey(settings.ACTION_TRAP_MOUSE, glfw.KeyEscape)
	input.BindActionKey(settings.ACTION_USE, glfw.KeyE)
	input.BindActionMouseMove(settings.ACTION_LOOK_HORZ, input.MOUSE_AXIS_X, settings.Current.MouseSensitivity)
	input.BindActionMouseMove(settings.ACTION_LOOK_VERT, input.MOUSE_AXIS_Y, settings.Current.MouseSensitivity)
	input.BindActionMouseButton(settings.ACTION_FIRE, glfw.MouseButton1)
	input.BindActionKey(settings.ACTION_SICKLE, glfw.Key1)
	input.BindActionKey(settings.ACTION_CHICKEN, glfw.Key2)
	input.BindActionKey(settings.ACTION_GRENADE, glfw.Key3)
	input.BindActionKey(settings.ACTION_PARUSU, glfw.Key4)
	// Double grenade
	// Sign of madness
	input.BindActionKey(settings.ACTION_AIRHORN, glfw.Key7)
	input.BindActionCharSequence(settings.ACTION_NOCLIP, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyC, glfw.KeyL, glfw.KeyI, glfw.KeyP})                               //TDCLIP
	input.BindActionCharSequence(settings.ACTION_GODMODE, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyD, glfw.KeyQ, glfw.KeyD})                                         //TDDQD
	input.BindActionCharSequence(settings.ACTION_MARYSUE, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyM, glfw.KeyS, glfw.KeyM})                                         //TDMSM
	input.BindActionCharSequence(settings.ACTION_DIE, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyU, glfw.KeyN, glfw.KeyA, glfw.KeyL, glfw.KeyI, glfw.KeyV, glfw.KeyE}) //TDUNALIVE
	input.BindActionCharSequence(settings.ACTION_KILL_ENEMIES, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyN, glfw.KeyU, glfw.KeyK, glfw.KeyE})                         //TDNUKE
	input.BindActionCharSequence(settings.ACTION_CAST_BLESSING, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyW, glfw.KeyO, glfw.KeyL, glfw.KeyO, glfw.KeyL, glfw.KeyO})  //TDWOLOLO

	mapName := settings.Current.Debug.StartMap
	if len(mapName) == 0 {
		mapName = "assets/maps/ti2-malicious-intents.te3"
	}

	app := &App{}
	app.LoadGame(mapName)
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
