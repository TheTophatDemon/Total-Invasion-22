package input

import "core:math"
import "vendor:glfw"

Capture_Callback :: proc "c" (new_binding: ^Binding, extra_data: rawptr)

capture_callback: Capture_Callback
capture_data: rawptr
captured_binding: Maybe(Binding)
frame_number: u64
last_key_pressed: Key
input_frame_number: u64
mouse_prev_x, mouse_prev_y: f64 = math.nan_f64(), math.nan_f64()
mouse_delta_x, mouse_delta_y: f64
mouse_scroll_y, mouse_prev_scroll_y: f32
anything_pressed: bool
was_any_mouse_button_pressed: bool

@(export)
init :: proc "c" () {
	glfw.SetKeyCallback(glfw.GetCurrentContext(), key_callback)
	glfw.SetMouseButtonCallback(glfw.GetCurrentContext(), mouse_callback)
	glfw.SetScrollCallback(glfw.GetCurrentContext(), scroll_callback)
}

@(export)
post_update :: proc "c" () {
	frame_number += 1
	anything_pressed = false
	was_any_mouse_button_pressed = any_mouse_button_pressed()
	last_key_pressed = .UNKNOWN
	mouse_pos_x, mouse_pos_y := glfw.GetCursorPos(glfw.GetCurrentContext())
	if !math.is_nan(mouse_prev_x) && !math.is_nan(mouse_prev_y) {
		mouse_delta_x = mouse_pos_x - mouse_prev_x
		mouse_delta_y = mouse_pos_y - mouse_prev_y
		moved_axis: Mouse_Axis = .NONE
		MOVE_THRESHOLD :: 5.0
		switch {
		case mouse_delta_x > MOVE_THRESHOLD:
			moved_axis = .POS_X
		case mouse_delta_x < -MOVE_THRESHOLD:
			moved_axis = .NEG_X
		case mouse_delta_y > MOVE_THRESHOLD:
			moved_axis = .POS_Y
		case mouse_delta_y < -MOVE_THRESHOLD:
			moved_axis = .NEG_Y
		}
		if moved_axis != .NONE && is_capturing_input() {
			captured_binding = new_mouse_movement_binding(moved_axis)
			end_capture_input()
		}
	}
	mouse_prev_x, mouse_prev_y = mouse_pos_x, mouse_pos_y
	mouse_prev_scroll_y = mouse_scroll_y
}

@(export)
trap_mouse :: proc "c" () {
	glfw.SetInputMode(glfw.GetCurrentContext(), glfw.CURSOR, glfw.CURSOR_DISABLED)
	mouse_prev_x, mouse_prev_y = glfw.GetCursorPos(glfw.GetCurrentContext())
}

@(export)
untrap_mouse :: proc "c" () {
	glfw.SetInputMode(glfw.GetCurrentContext(), glfw.CURSOR, glfw.CURSOR_NORMAL)
}

@(export)
is_mouse_trapped :: proc "c" () -> bool {
	return glfw.GetInputMode(glfw.GetCurrentContext(), glfw.CURSOR) == glfw.CURSOR_DISABLED
}

@(export)
capture_input :: proc "c" (callback: Capture_Callback, extra_data: rawptr) {
	capture_callback = callback
	capture_data = extra_data
}

@(export)
is_capturing_input :: proc "c" () -> bool {
	return capture_callback != nil
}

@(export)
capture_extra_data :: proc "c" () -> rawptr {
	if capture_callback != nil {
		return capture_data
	}
	return nil
}

end_capture_input :: proc "c" () {
	if capture_callback != nil {
		capture_callback(&captured_binding.?, capture_data)
		capture_callback = nil
		capture_data = nil
		captured_binding = nil
	}
}

@(export)
mouse_position :: proc "c" () -> [2]f32 {
	if glfw.GetWindowAttrib(glfw.GetCurrentContext(), glfw.FOCUSED) != 1 {
		return [2]f32{-10000.0, -10000.0}
	}
	x, y := glfw.GetCursorPos(glfw.GetCurrentContext())
	return [2]f32{f32(x), f32(y)}
}

@(export)
mouse_delta :: proc "c" () -> [2]f32 {
	return [2]f32{f32(mouse_delta_x), f32(mouse_delta_y)}
}

@(export)
mouse_scroll :: proc "c" () -> f32 {
	return mouse_scroll_y
}

@(export)
mouse_scroll_delta :: proc "c" () -> f32 {
	return mouse_scroll_y - mouse_prev_scroll_y
}

@(export)
set_mouse_position :: proc "c" (x, y: f32) {
	mouse_prev_x = f64(x)
	mouse_prev_y = f64(y)
	mouse_delta_x, mouse_delta_y = 0, 0
	glfw.SetCursorPos(glfw.GetCurrentContext(), mouse_prev_x, mouse_prev_y)
}

@(export)
is_mouse_button_down :: proc "c" (button: Mouse_Button) -> bool {
	return glfw.GetMouseButton(glfw.GetCurrentContext(), i32(button)) == glfw.PRESS && !is_capturing_input()
}

@(export)
any_mouse_button_pressed :: proc "c" () -> bool {
	return is_mouse_button_down(.LEFT) || is_mouse_button_down(.RIGHT) || is_mouse_button_down(.MIDDLE)
}

@(export)
any_mouse_button_just_pressed :: proc "c" () -> bool {
	return any_mouse_button_pressed() && !was_any_mouse_button_pressed && !is_capturing_input()
}

@(export)
is_anything_pressed :: proc "c" () -> bool {
	return anything_pressed && !is_capturing_input()
}

@(private)
key_callback :: proc "c" (window: glfw.WindowHandle, key, scancode, action, mods: i32) {
	switch action {
		case glfw.PRESS:
			if is_capturing_input() {
				captured_binding = new_key_binding(Key(key))
			} else {
				anything_pressed = true
				last_key_pressed = Key(key)
			}
		case glfw.RELEASE:
			if is_capturing_input() && captured_binding != nil {
				end_capture_input()
			}
	}
}

@(private)
mouse_callback :: proc "c" (window: glfw.WindowHandle, button, action, mods: i32) {
	switch action {
		case glfw.PRESS:
			if is_capturing_input() {
				captured_binding = new_mouse_button_binding(Mouse_Button(button))
			} else {
				anything_pressed = true
			}
		case glfw.RELEASE:
			if is_capturing_input() && captured_binding != nil {
				end_capture_input()
			}
	}
}

@(private)
scroll_callback :: proc "c" (window: glfw.WindowHandle, xoffset, yoffset: f64) {
	mouse_scroll_y = f32(yoffset)
}