package tdaudio

import (
	"log"
	"structs"

	"github.com/ebitengine/purego"
	"tophatdemon.com/total-invasion-ii/engine/failure"
)

type (
	SoundId uint32
	VoiceId struct {
		_             structs.HostLayout
		Player        SoundId
		Id, PlayCount uint32
	}
)

var (
	loadSound              func(path string, polyphony uint8, looping bool, rolloff float32, noRetrigger bool) SoundId
	isLooped               func(soundId SoundId) bool
	playSound              func(soundId SoundId, x, y, z float32, attenuate bool) VoiceId
	isPlaying              func(voiceId VoiceId) bool
	setSoundPosition       func(voiceId VoiceId, x, y, z float32)
	voiceIsValid           func(voiceId VoiceId) bool
	stopSound              func(voiceId VoiceId)
	seekSound              func(voiceId VoiceId, ms uint64)
	getSoundTime           func(voiceId VoiceId) uint64
	initAudio              func() bool
	setListenerOrientation func(posX, posY, posZ, dirX, dirY, dirZ float32)
	setSfxVolume           func(volume float32)
	getSfxVolume           func() float32
	setMusicVolume         func(volume float32)
	getMusicVolume         func() float32
	queueSong              func(path string, looping bool, fadeoutMs uint64) bool
	stopAllSounds          func()
	update                 func()
	teardown               func()
	freeSounds             func()
)

// Sound functions

func LoadSound(path string, polyphony uint8, looping bool, rolloff float32, noRetrigger bool) SoundId {
	log.Println("Loading sound at ", path)
	cSound := loadSound(path, polyphony, looping, rolloff, noRetrigger)
	return SoundId(cSound)
}

func (sound SoundId) IsLooping() bool {
	return isLooped(sound)
}

func (sound SoundId) Play() VoiceId {
	return playSound(sound, 0, 0, 0, false)
}

func (sound SoundId) PlayAttenuated(x, y, z float32) VoiceId {
	return playSound(sound, x, y, z, true)
}

func (sound SoundId) PlayAttenuatedV(pos [3]float32) VoiceId {
	return sound.PlayAttenuated(pos[0], pos[1], pos[2])
}

func (sound SoundId) IsValid() bool {
	return int(sound) != 0
}

// Voice functions

func (voice VoiceId) IsPlaying() bool {
	return isPlaying(voice)
}

func (voice VoiceId) SetPosition(x, y, z float32) {
	setSoundPosition(voice, x, y, z)
}

func (voice VoiceId) SetPositionV(pos [3]float32) {
	voice.SetPosition(pos[0], pos[1], pos[2])
}

func (voice VoiceId) IsValid() bool {
	return voiceIsValid(voice)
}

func (voice VoiceId) Stop() {
	stopSound(voice)
}

func (voice VoiceId) Seek(timeMs uint64) {
	seekSound(voice, timeMs)
}

// Returns the current audio track time in milliseconds.
func (voice VoiceId) GetTime() uint64 {
	return getSoundTime(voice)
}

func (voice VoiceId) Sound() SoundId {
	return SoundId(voice.Player)
}

func Init() bool {
	lib, err := openLibrary()
	if err != nil {
		failure.LogErrWithLocation("error loading TD audio dll: %v", err)
		return false
	}
	purego.RegisterLibFunc(&loadSound, lib, "load_sound")
	purego.RegisterLibFunc(&isLooped, lib, "sound_is_looped")
	purego.RegisterLibFunc(&playSound, lib, "play_sound")
	purego.RegisterLibFunc(&isPlaying, lib, "sound_is_playing")
	purego.RegisterLibFunc(&setSoundPosition, lib, "sound_set_position")
	purego.RegisterLibFunc(&voiceIsValid, lib, "voice_is_valid")
	purego.RegisterLibFunc(&stopSound, lib, "stop_sound")
	purego.RegisterLibFunc(&seekSound, lib, "seek_sound")
	purego.RegisterLibFunc(&getSoundTime, lib, "get_sound_time")
	purego.RegisterLibFunc(&initAudio, lib, "audio_init")
	purego.RegisterLibFunc(&setListenerOrientation, lib, "set_listener_orientation")
	purego.RegisterLibFunc(&setSfxVolume, lib, "set_sfx_volume")
	purego.RegisterLibFunc(&getSfxVolume, lib, "get_sfx_volume")
	purego.RegisterLibFunc(&setMusicVolume, lib, "set_music_volume")
	purego.RegisterLibFunc(&getMusicVolume, lib, "get_music_volume")
	purego.RegisterLibFunc(&queueSong, lib, "queue_song")
	purego.RegisterLibFunc(&stopAllSounds, lib, "stop_all_sounds")
	purego.RegisterLibFunc(&update, lib, "audio_update")
	purego.RegisterLibFunc(&teardown, lib, "audio_teardown")
	purego.RegisterLibFunc(&freeSounds, lib, "free_sounds")
	return initAudio()
}

func SetListenerOrientation(posX, posY, posZ, dirX, dirY, dirZ float32) {
	setListenerOrientation(posX, posY, posZ, dirX, dirY, dirZ)
}

func SetListenerOrientationV(pos, dir [3]float32) {
	SetListenerOrientation(pos[0], pos[1], pos[2], dir[0], dir[1], dir[2])
}

func SetSfxVolume(newVolume float32) {
	setSfxVolume(newVolume)
}

func SfxVolume() float32 {
	return getSfxVolume()
}

func SetMusicVolume(newVolume float32) {
	setMusicVolume(newVolume)
}

func MusicVolume() float32 {
	return getMusicVolume()
}

func QueueSong(path string, looping bool, fadeoutMillis uint64) bool {
	if len(path) == 0 {
		return queueSong("", looping, fadeoutMillis)
	}
	log.Println("Loading song at ", path)
	return queueSong(path, looping, fadeoutMillis)
}

func StopAll() {
	stopAllSounds()
}

func Update() {
	update()
}

func FreeSounds() {
	freeSounds()
}

func Teardown() {
	teardown()
}
