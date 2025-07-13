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
		app.LoadGame(msg.NextMapPath, msg)
	}
}

func (app *App) LoadGame(mapPath string, sig game.MapChangeSignal) {
	log.Println("Loading game at map ", mapPath)

	cache.Reset()
	cache.DefaultFont, _ = cache.GetFont("assets/textures/ui/font.fnt")

	world, err := world.NewWorld(app, mapPath, sig)
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
	input.BindActionCharSequence(settings.ActionKillEnemies, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyN, glfw.KeyU, glfw.KeyK, glfw.KeyE})                          //TDNUKE
	input.BindActionCharSequence(settings.ActionCastBlessing, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyW, glfw.KeyO, glfw.KeyL, glfw.KeyO, glfw.KeyL, glfw.KeyO})   //TDWOLOLO

	mapName := settings.Current.Debug.StartMap
	if len(mapName) == 0 {
		mapName = "assets/maps/ti2-malicious-intents.te3"
	}

	app := &App{}
	app.LoadGame(mapName, game.MapChangeSignal{
		NextMapPath: mapName,
	})
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
