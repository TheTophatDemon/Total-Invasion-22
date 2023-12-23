package settings

import "tophatdemon.com/total-invasion-ii/engine/input"

const (
	WINDOW_WIDTH        = 1280
	WINDOW_HEIGHT       = 720
	WINDOW_ASPECT_RATIO = float32(WINDOW_WIDTH) / WINDOW_HEIGHT
)

const (
	ACTION_FORWARD input.Action = iota
	ACTION_BACK
	ACTION_LEFT
	ACTION_RIGHT
	ACTION_SLOW
	ACTION_LOOK_HORZ
	ACTION_LOOK_VERT
	ACTION_TRAP_MOUSE
	ACTION_FIRE
	ACTION_USE
	ACTION_NOCLIP
	ACTION_COUNT
)

var actionNames = [ACTION_COUNT]string{
	ACTION_FORWARD:    "Move Forward",
	ACTION_BACK:       "Move Back",
	ACTION_LEFT:       "Strafe Left",
	ACTION_RIGHT:      "Strafe Right",
	ACTION_SLOW:       "Slow",
	ACTION_LOOK_HORZ:  "Look Horizontally",
	ACTION_LOOK_VERT:  "Look Vertically",
	ACTION_TRAP_MOUSE: "Trap Mouse",
	ACTION_FIRE:       "Fire",
	ACTION_USE:        "Use",
	ACTION_NOCLIP:     "Noclip",
}

const (
	MOUSE_SENSITIVITY = 0.005
)

func ActionName(action input.Action) string {
	if action > ACTION_COUNT {
		return ""
	}
	return actionNames[action]
}
