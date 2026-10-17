package input

import "core:math"
import "vendor:glfw"

MOUSE_DEAD_ZONE :: 0.05

Key :: enum(i32) {
	UNKNOWN = -1,
	SPACE = 32,
	APOSTROPHE = 39,
	COMMA = 44,
	MINUS = 45,
	PERIOD = 46,
	SLASH = 47,
	SEMICOLON = 59,
	EQUAL = 61,
	LEFT_BRACKET = 91,
	BACKSLASH = 92,
	RIGHT_BRACKET = 93,
	GRAVE_ACCENT = 96,
	WORLD_1 = 161,
	WORLD_2 = 162,
	ZERO = 48,
	ONE = 49,
	TWO = 50,
	THREE = 51,
	FOUR = 52,
	FIVE = 53,
	SIX = 54,
	SEVEN = 55,
	EIGHT = 56,
	NINE = 57,
	A = 65,
	B = 66,
	C = 67,
	D = 68,
	E = 69,
	F = 70,
	G = 71,
	H = 72,
	I = 73,
	J = 74,
	K = 75,
	L = 76,
	M = 77,
	N = 78,
	O = 79,
	P = 80,
	Q = 81,
	R = 82,
	S = 83,
	T = 84,
	U = 85,
	V = 86,
	W = 87,
	X = 88,
	Y = 89,
	Z = 90,
	ESCAPE = 256,
	ENTER = 257,
	TAB = 258,
	BACKSPACE = 259,
	INSERT = 260,
	DELETE = 261,
	RIGHT = 262,
	LEFT = 263,
	DOWN = 264,
	UP = 265,
	PAGE_UP = 266,
	PAGE_DOWN = 267,
	HOME = 268,
	END = 269,
	CAPS_LOCK = 280,
	SCROLL_LOCK = 281,
	NUM_LOCK = 282,
	PRINT_SCREEN = 283,
	PAUSE = 284,
	F1 = 290,
	F2 = 291,
	F3 = 292,
	F4 = 293,
	F5 = 294,
	F6 = 295,
	F7 = 296,
	F8 = 297,
	F9 = 298,
	F10 = 299,
	F11 = 300,
	F12 = 301,
	F13 = 302,
	F14 = 303,
	F15 = 304,
	F16 = 305,
	F17 = 306,
	F18 = 307,
	F19 = 308,
	F20 = 309,
	F21 = 310,
	F22 = 311,
	F23 = 312,
	F24 = 313,
	F25 = 314,
	KP_0 = 320,
	KP_1 = 321,
	KP_2 = 322,
	KP_3 = 323,
	KP_4 = 324,
	KP_5 = 325,
	KP_6 = 326,
	KP_7 = 327,
	KP_8 = 328,
	KP_9 = 329,
	KP_DECIMAL = 330,
	KP_DIVIDE = 331,
	KP_MULTIPLY = 332,
	KP_SUBTRACT = 333,
	KP_ADD = 334,
	KP_ENTER = 335,
	KP_EQUAL = 336,
	LEFT_SHIFT = 340,
	LEFT_CONTROL = 341,
	LEFT_ALT = 342,
	LEFT_SUPER = 343,
	RIGHT_SHIFT = 344,
	RIGHT_CONTROL = 345,
	RIGHT_ALT = 346,
	RIGHT_SUPER = 347,
	MENU = 348,
}

Mouse_Button :: enum(i32) {
	LEFT,
	RIGHT,
	MIDDLE,
	FOUR,
	FIVE,
	SIX,
	SEVEN,
	EIGHT,
}

Mouse_Axis :: enum(u8) {
	NONE,
	POS_X,
	NEG_X,
	POS_Y,
	NEG_Y,
}

Key_Binding :: struct {
	key: Key,
}

Char_Sequence_Binding :: struct {
	sequence: [dynamic; 16]Key,
	progress: int,
}

Mouse_Button_Binding :: struct {
	mouse_button: Mouse_Button,
}

Mouse_Movement_Binding :: struct {
	mouse_axis: Mouse_Axis,
}

Binding :: struct {
	was_pressed, just_pressed, just_released: bool,
	last_updated_frame: u64,
	variant: union #no_nil {
		Key_Binding,
		Char_Sequence_Binding,
		Mouse_Button_Binding,
		Mouse_Movement_Binding,
	}
}

binding_pressed :: proc "c" (binding: ^Binding) -> bool {
	is_pressed: bool
	switch &v in binding.variant {
		case Key_Binding:
			is_pressed = glfw.GetKey(glfw.GetCurrentContext(), i32(v.key)) == glfw.PRESS
		case Char_Sequence_Binding:
			if last_key_pressed != .UNKNOWN && binding.last_updated_frame != frame_number {
				if v.progress == len(v.sequence) {
					v.progress = 0
				}
				if v.progress < len(v.sequence) && last_key_pressed == v.sequence[v.progress] {
					v.progress += 1
				} else {
					v.progress = 0
				}
				binding.last_updated_frame = frame_number
			}
			is_pressed = (v.progress == len(v.sequence))
		case Mouse_Button_Binding:
			is_pressed = glfw.GetMouseButton(glfw.GetCurrentContext(), i32(v.mouse_button)) == glfw.PRESS
		case Mouse_Movement_Binding:
			is_pressed = (math.abs(binding_axis(binding)) > MOUSE_DEAD_ZONE)
	}
	
	if is_capturing_input() {
		binding.just_pressed, binding.just_released, binding.was_pressed = false, false, false
		return false
	}
	if binding.last_updated_frame != frame_number {
		binding.just_pressed = is_pressed && !binding.was_pressed
		binding.just_released = !is_pressed && binding.was_pressed
		binding.was_pressed = is_pressed
		binding.last_updated_frame = frame_number
	}
	return is_pressed
}

binding_just_pressed :: proc "c" (binding: ^Binding) -> bool {
	binding_pressed(binding)
	return binding.just_pressed
}

binding_just_released :: proc "c" (binding: ^Binding) -> bool {
	binding_pressed(binding)
	return binding.just_released
}

binding_axis :: proc "c" (binding: ^Binding) -> f32 {
	switch v in binding.variant {
		case Key_Binding, Mouse_Button_Binding:
			if binding_pressed(binding) {
				return 1.0
			} else {
				return 0.0
			}
		case Char_Sequence_Binding:
			return f32(v.progress) / f32(len(v.sequence))
		case Mouse_Movement_Binding:
			scaled_x := mouse_delta_x / 10.0
			scaled_y := mouse_delta_y / 10.0
			switch v.mouse_axis {
				case .NONE:
					return 0.0
				case .POS_X:
					return max(0.0, f32(scaled_x))
				case .NEG_X:
					return -min(0.0, f32(scaled_x))
				case .POS_Y:
					return max(0.0, f32(scaled_y))
				case .NEG_Y:
					return -min(0.0, f32(scaled_y))
			}
			return 0.0
	}
	return 0.0
}

new_key_binding :: proc "c" (key: Key) -> Binding {
	return Binding{
		variant = Key_Binding{
			key = key,
		},
	}
}

new_mouse_button_binding :: proc "c" (mouse_button: Mouse_Button) -> Binding {
	return Binding{
		variant = Mouse_Button_Binding {
			mouse_button = mouse_button,
		},
	}
}

new_mouse_movement_binding :: proc "c" (axis: Mouse_Axis) -> Binding {
	return Binding{
		variant = Mouse_Movement_Binding{
			mouse_axis = axis,
		},
	}
}

new_char_sequence_binding :: proc "c" (sequence: []Key) -> Binding {
	char_seq := Char_Sequence_Binding{
		sequence = {},
	}
	for key in sequence {
		append(&char_seq.sequence, key)
	}
	return Binding{
		variant = char_seq,
	}
}