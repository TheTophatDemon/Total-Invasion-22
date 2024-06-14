package main

import (
	"log"
	"os"
	"runtime"
	"runtime/pprof"

	"github.com/go-gl/glfw/v3.3/glfw"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/input"

	"tophatdemon.com/total-invasion-ii/game/settings"
	"tophatdemon.com/total-invasion-ii/game/world"
)

type Game struct {
	world *world.World
}

func (game *Game) Update(deltaTime float32) {
	// Free mouse
	if input.IsActionJustPressed(settings.ACTION_TRAP_MOUSE) {
		if input.IsMouseTrapped() {
			input.UntrapMouse()
		} else {
			input.TrapMouse()
		}
	}

	if input.IsActionJustPressed(settings.ACTION_MUTE_MUS) {
		if settings.Current.MusicVolume() != 0.0 {
			settings.UpdateMusicVolume(0.0)
		} else {
			settings.UpdateMusicVolume(settings.Default.MusicVolume())
		}
	}

	game.world.Update(deltaTime)
}

func (game *Game) Render() {
	game.world.Render()
}

func main() {
	cpuProfile, err := os.Create("cpuProfile.pprof")
	if err != nil {
		log.Fatal(err)
	}
	defer cpuProfile.Close()
	if err := pprof.StartCPUProfile(cpuProfile); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	err = engine.Init(int(settings.Current.WindowWidth), int(settings.Current.WindowHeight), "Total Invasion 22")
	defer engine.DeInit()
	if err != nil {
		panic(err)
	}

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
	input.BindActionCharSequence(settings.ACTION_NOCLIP, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyC, glfw.KeyL, glfw.KeyI, glfw.KeyP})
	input.BindActionKey(settings.ACTION_MUTE_MUS, glfw.KeyM)

	world, err := world.NewWorld("assets/maps/ti2-malicious-intents.te3")
	if err != nil {
		panic(err)
	}

	settings.UpdateMusicVolume(0.0)

	runtime.GC()

	input.TrapMouse()
	engine.Run(&Game{
		world,
	})

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
