package lib

import "core:slice"
import "input"
import "audio"

//AUDIO

@(export)
audio_free_sounds :: proc "c" () {
	audio.free_sounds()
}

@(export)
audio_voice_is_valid :: proc "c" (voice_id: audio.Voice_Id) -> bool {
	return audio.voice_is_valid(voice_id)
}

@(export)
audio_load_sound :: proc "c" (path: cstring, polyphony: u8, looping: bool, rolloff: f32, no_retrigger: bool) -> audio.Player_Id {
	return audio.load_sound(path, polyphony, looping, rolloff, no_retrigger)
}

@(export)
audio_sound_is_looped :: proc "c" (player_id: audio.Player_Id) -> bool {
	return audio.sound_is_looped(player_id)
}

@(export)
audio_init :: proc "c" () -> bool {
	return audio.audio_init()
}

@(export)
audio_play_sound :: proc "c" (player_id: audio.Player_Id, x, y, z: f32, attenuated: bool) -> audio.Voice_Id {
	return audio.play_sound(player_id, x, y, z, attenuated)
}

@(export)
audio_sound_is_playing :: proc "c" (voice: audio.Voice_Id) -> bool {
	return audio.sound_is_playing(voice)
}

@(export)
audio_sound_set_position :: proc "c" (voice: audio.Voice_Id, x, y, z: f32) {
	audio.sound_set_position(voice, x, y, z)
}

@(export)
audio_stop_sound :: proc "c" (voice: audio.Voice_Id) {
	audio.stop_sound(voice)
}

@(export)
audio_stop_all_sounds :: proc "c" () {
	audio.stop_all_sounds()
}

@(export)
audio_seek_sound :: proc "c" (voice: audio.Voice_Id, time_ms: u64) {
	audio.seek_sound(voice, time_ms)
}

@(export)
audio_get_sound_time :: proc "c" (voice: audio.Voice_Id) -> u64 {
	return audio.get_sound_time(voice)
}

@(export)
audio_set_listener_orientation :: proc "c" (pos_x, pos_y, pos_z, dir_x, dir_y, dir_z: f32) {
	audio.set_listener_orientation(pos_x, pos_y, pos_z, dir_x, dir_y, dir_z)
}

@(export)
audio_set_sfx_volume :: proc "c" (new_volume: f32) {
	audio.set_sfx_volume(new_volume)
}

@(export)
audio_get_sfx_volume :: proc "c" () -> f32 {
	return audio.get_sfx_volume()
}

@(export)
audio_set_music_volume :: proc "c" (new_volume: f32) {
	audio.set_music_volume(new_volume)
}

@(export)
audio_get_music_volume :: proc "c" () -> f32 {
	return audio.get_music_volume()
}

@(export)
audio_queue_song :: proc "c" (path: cstring, looping: bool, fadeout_millis: u64) -> bool {
	return audio.queue_song(path, looping, fadeout_millis)
}

@(export)
audio_update :: proc "c" () {
	audio.audio_update()
} 

@(export)
audio_teardown :: proc "c" () {
	audio.audio_teardown()
}

//INPUT

@(export)
input_init :: proc "c" () {
	input.init()
}

@(export)
input_post_update :: proc "c" () {
	input.post_update()
}

@(export)
input_trap_mouse :: proc "c" () {
	input.trap_mouse()
}

@(export)
input_untrap_mouse :: proc "c" () {
	input.untrap_mouse()
}

@(export)
input_is_mouse_trapped :: proc "c" () -> bool {
	return input.is_mouse_trapped()
}

@(export)
input_capture_input :: proc "c" (callback: input.Capture_Callback, extra_data: rawptr) {
	input.capture_input(callback, extra_data)
}

@(export)
input_is_capturing_input :: proc "c" () -> bool {
	return input.is_capturing_input()
}

@(export)
input_capture_extra_data :: proc "c" () -> rawptr {
	return input.capture_extra_data()
}

@(export)
input_mouse_position :: proc "c" () -> [2]f32 {
	return input.mouse_position()
}

@(export)
input_mouse_delta :: proc "c" () -> [2]f32 {
	return input.mouse_delta()
}

@(export)
input_mouse_scroll :: proc "c" () -> f32 {
	return input.mouse_scroll()
}

@(export)
input_mouse_scroll_delta :: proc "c" () -> f32 {
	return input.mouse_scroll_delta()
}

@(export)
input_set_mouse_position :: proc "c" (x, y: f32) {
	input.set_mouse_position(x, y)
}

@(export)
input_is_mouse_button_down :: proc "c" (button: input.Mouse_Button) -> bool {
	return input.is_mouse_button_down(button)
}

@(export)
input_any_mouse_button_pressed :: proc "c" () -> bool {
	return input.any_mouse_button_pressed()
}

@(export)
input_any_mouse_button_just_pressed :: proc "c" () -> bool {
	return input.any_mouse_button_just_pressed()
}

@(export)
input_is_anything_pressed :: proc "c" () -> bool {
	return input.is_anything_pressed()
}

@(export)
input_binding_pressed :: proc "c" (binding: ^input.Binding) -> bool {
	return input.binding_pressed(binding)
}

@(export)
input_binding_just_pressed :: proc "c" (binding: ^input.Binding) -> bool {
	return input.binding_just_pressed(binding)
}

@(export)
input_binding_just_released :: proc "c" (binding: ^input.Binding) -> bool {
	return input.binding_just_released(binding)
}

@(export)
input_binding_axis :: proc "c" (binding: ^input.Binding) -> f32 {
	return input.binding_axis(binding)
}

@(export)
input_new_key_binding :: proc "c" (key: input.Key) -> input.Binding {
	return input.new_key_binding(key)
}

@(export)
input_new_mouse_button_binding :: proc "c" (mouse_button: input.Mouse_Button) -> input.Binding {
	return input.new_mouse_button_binding(mouse_button)
}

@(export)
input_new_mouse_movement_binding :: proc "c" (axis: input.Mouse_Axis) -> input.Binding {
	return input.new_mouse_movement_binding(axis)
}

@(export)
input_new_char_sequence_binding :: proc "c" (keys: [^]input.Key, keys_count: int) -> input.Binding {
	return input.new_char_sequence_binding(slice.from_ptr(keys, keys_count))
}