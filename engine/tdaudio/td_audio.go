package tdaudio

/*
#include "./td_audio.h"
#include <stdlib.h>
*/
import "C"
import (
	"log"
	"unsafe"
)

type (
	SoundId C.td_player_id
	VoiceId C.td_voice_id
)

// Sound functions

func LoadSound(path string, polyphony uint8, looping bool, rolloff float32) SoundId {
	log.Println("Loading sound at ", path)
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	cSound := C.td_audio_load_sound(cPath, C.uint8_t(polyphony), C.bool(looping), C.float(rolloff))
	return SoundId(cSound)
}

func (sound SoundId) IsLooping() bool {
	return bool(C.td_audio_sound_is_looped(C.td_player_id(sound)))
}

func (sound SoundId) Play() VoiceId {
	return VoiceId(C.td_audio_play_sound(C.td_player_id(sound), C.float(0), C.float(0), C.float(0), C.bool(false)))
}

func (sound SoundId) PlayAttenuated(x, y, z float32) VoiceId {
	return VoiceId(C.td_audio_play_sound(C.td_player_id(sound), C.float(x), C.float(y), C.float(z), C.bool(true)))
}

func (sound SoundId) PlayAttenuatedV(pos [3]float32) VoiceId {
	return sound.PlayAttenuated(pos[0], pos[1], pos[2])
}

// Voice functions

func (voice VoiceId) IsPlaying() bool {
	return bool(C.td_audio_sound_is_playing(C.td_voice_id(voice)))
}

func (voice VoiceId) SetPosition(x, y, z float32) {
	C.td_audio_set_sound_position(C.td_voice_id(voice), C.float(x), C.float(y), C.float(z))
}

func (voice VoiceId) SetPositionV(pos [3]float32) {
	voice.SetPosition(pos[0], pos[1], pos[2])
}

func (voice VoiceId) IsValid() bool {
	return bool(C.td_audio_voice_is_valid(C.td_voice_id(voice)))
}

func (voice VoiceId) Stop() {
	C.td_audio_stop_sound(C.td_voice_id(voice))
}

// General audio functions

func Init() bool {
	return bool(C.td_audio_init())
}

func SetListenerOrientation(posX, posY, posZ, dirX, dirY, dirZ float32) {
	C.td_audio_set_listener_orientation(C.float(posX), C.float(posY), C.float(posZ), C.float(dirX), C.float(dirY), C.float(dirZ))
}

func SetListenerOrientationV(pos, dir [3]float32) {
	SetListenerOrientation(pos[0], pos[1], pos[2], dir[0], dir[1], dir[2])
}

func SetSfxVolume(newVolume float32) {
	C.td_audio_set_sfx_volume(C.float(newVolume))
}

func SfxVolume() float32 {
	return float32(C.td_audio_get_sfx_volume())
}

func SetMusicVolume(newVolume float32) {
	C.td_audio_set_music_volume(C.float(newVolume))
}

func MusicVolume() float32 {
	return float32(C.td_audio_get_music_volume())
}

func QueueSong(path string, looping bool, fadeoutMillis uint64) bool {
	if len(path) == 0 {
		return bool(C.td_audio_queue_song(nil, C.bool(looping), C.uint64_t(fadeoutMillis)))
	}
	log.Println("Loading song at ", path)
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	return bool(C.td_audio_queue_song(cPath, C.bool(looping), C.uint64_t(fadeoutMillis)))
}

func Update() {
	C.td_audio_update()
}

func FreeSounds() {
	C.td_audio_free_sounds()
}

func Teardown() {
	C.td_audio_teardown()
}
