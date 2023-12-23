package main

import (
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"

	"tophatdemon.com/total-invasion-ii/engine"
	"tophatdemon.com/total-invasion-ii/engine/input"

	"tophatdemon.com/total-invasion-ii/game"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type Game struct {
	world *game.World
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

	game.world.Update(deltaTime)
}

func (game *Game) Render() {
	game.world.Render()
}

func main() {
	err := engine.Init(settings.WINDOW_WIDTH, settings.WINDOW_HEIGHT, "Total Invasion II")
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
	input.BindActionMouseMove(settings.ACTION_LOOK_HORZ, input.MOUSE_AXIS_X, settings.MOUSE_SENSITIVITY)
	input.BindActionMouseMove(settings.ACTION_LOOK_VERT, input.MOUSE_AXIS_Y, settings.MOUSE_SENSITIVITY)
	input.BindActionMouseButton(settings.ACTION_FIRE, glfw.MouseButton1)
	input.BindActionCharSequence(settings.ACTION_NOCLIP, []glfw.Key{glfw.KeyT, glfw.KeyD, glfw.KeyC, glfw.KeyL, glfw.KeyI, glfw.KeyP})

	input.TrapMouse()
	world, err := game.NewWorld("assets/maps/ti2-malicious-intents.te3")
	if err != nil {
		panic(err)
	}

	runtime.GC()

	engine.Run(&Game{
		world,
	})
}
