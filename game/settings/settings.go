package settings

const (
	WINDOW_WIDTH        = 1280
	WINDOW_HEIGHT       = 720
	WINDOW_ASPECT_RATIO = float32(WINDOW_WIDTH) / WINDOW_HEIGHT

	ACTION_FORWARD    = "MoveForward"
	ACTION_BACK       = "MoveBack"
	ACTION_LEFT       = "StrafeLeft"
	ACTION_RIGHT      = "StrafeRight"
	ACTION_LOOK_HORZ  = "LookHorz"
	ACTION_LOOK_VERT  = "LookVert"
	ACTION_TRAP_MOUSE = "TrapMouse"

	MOUSE_SENSITIVITY = 0.005
)
