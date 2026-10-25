package audio

/*
#define STB_VORBIS_HEADER_ONLY
#include "./stb_vorbis.c"
#undef STB_VORBIS_HEADER_ONLY
#define MA_NO_MP3
#define MA_NO_FLAC
#define MA_NO_ENCODING
#define MINIAUDIO_IMPLEMENTATION
#include "./miniaudio.h"

ma_engine g_engine;

void data_callback(ma_device* pDevice, void* pOutput, const void* pInput, ma_uint32 frameCount)
{
    ma_engine_read_pcm_frames(&g_engine, pOutput, frameCount, NULL);
}

*/
import "C"
import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/failure"
)

const (
	SampleRate   = 44100
	ChannelCount = 2

	LoopStartTag  = "LOOPSTART="
	LoopLengthTag = "LOOPLENGTH="
)

type (
	Voice struct {
		sound     C.ma_sound
		playCount uint32
	}
	Player struct {
		voices      []Voice
		noRetrigger bool
	}
)

var (
	device                *C.ma_device
	resourceManager       *C.ma_resource_manager
	sfxGroup, musicGroup  *C.ma_sound_group
	songs                 [2]Voice
	currentSong, nextSong *Voice
	systemsPinner         runtime.Pinner
)

func Init() bool {
	// Resource manager init
	{
		config := C.ma_resource_manager_config_init()
		config.decodedFormat = C.ma_format_f32
		config.decodedChannels = ChannelCount
		config.decodedSampleRate = SampleRate
		systemsPinner.Pin(&config)
		resourceManager = new(C.ma_resource_manager)
		systemsPinner.Pin(resourceManager)
		if result := C.ma_resource_manager_init(&config, resourceManager); result != C.MA_SUCCESS {
			failure.LogErrWithLocation("failed to initialize mini audio resource manager, code %d", result)
			return false
		}
	}

	// Device init
	{
		config := C.ma_device_config_init(C.ma_device_type_playback)
		config.playback.format = C.ma_format_f32
		config.playback.channels = ChannelCount
		config.sampleRate = SampleRate
		config.dataCallback = C.ma_device_data_proc(C.data_callback)
		systemsPinner.Pin(&config)
		device = new(C.ma_device)
		systemsPinner.Pin(device)
		if result := C.ma_device_init(nil, &config, device); result != C.MA_SUCCESS {
			failure.LogErrWithLocation("failed to initialize audio system, code %d", result)
			return false
		}
	}

	// Engine init
	{
		config := C.ma_engine_config_init()
		config.pDevice = device
		config.pResourceManager = resourceManager
		configPtr := &config
		systemsPinner.Pin(configPtr)
		if result := C.ma_engine_init(configPtr, &C.g_engine); result != C.MA_SUCCESS {
			failure.LogErrWithLocation("failed to initialize miniaudio engine, code %d", result)
			return false
		}
	}

	// Initialize sound groups
	sfxGroup = new(C.ma_sound_group)
	systemsPinner.Pin(sfxGroup)
	if result := C.ma_sound_group_init(&C.g_engine, 0, nil, sfxGroup); result != C.MA_SUCCESS {
		failure.LogErrWithLocation("failed to initialize sfx group, code %d", result)
		return false
	}
	musicGroup = new(C.ma_sound_group)
	systemsPinner.Pin(musicGroup)
	if result := C.ma_sound_group_init(&C.g_engine, 0, nil, musicGroup); result != C.MA_SUCCESS {
		failure.LogErrWithLocation("failed to initialize music group, code %d", result)
		return false
	}

	// Start the device
	if result := C.ma_device_start(device); result != C.MA_SUCCESS {
		failure.LogErrWithLocation("failed to start playback device, code %d", result)
		return false
	}

	return true
}

type SfxOptions struct {
	Loop        bool
	Polyphony   int
	Rolloff     float32
	NoRetrigger bool
}

func LoadSfx(soundPath string) (*Player, error) {
	if !strings.HasSuffix(soundPath, ".wav") {
		return nil, fmt.Errorf("%v is not a wav file", soundPath)
	}

	// Look for metadata file
	metaPath := strings.TrimSuffix(soundPath, ".wav") + ".toml"
	metadata, err := assets.LoadAndUnmarshalTOML[SfxOptions](metaPath)
	if _, ok := err.(*os.PathError); err != nil && !ok {
		// The file is optional, so print errors that aren't 'file not found'.
		return nil, fmt.Errorf("could not parse metadata for %s; %s", soundPath, err)
	}

	polyphony := uint8(4)
	looped := false
	rolloff := float32(0.1)
	noRetrig := false
	if metadata != nil {
		if metadata.Polyphony != 0 {
			polyphony = uint8(metadata.Polyphony)
		}
		looped = metadata.Loop
		if metadata.Rolloff != 0.0 {
			rolloff = metadata.Rolloff
		}
		noRetrig = metadata.NoRetrigger
	}

	player := new(Player{
		noRetrigger: noRetrig,
		voices:      make([]Voice, polyphony),
	})
	systemsPinner.Pin(player)

	cPath := C.CString(soundPath)
	defer C.free(unsafe.Pointer(cPath))

	for v := range player.voices {
		sound := &player.voices[v].sound
		flags := C.MA_SOUND_FLAG_DECODE
		result := C.ma_sound_init_from_file(&C.g_engine, cPath, C.ma_uint32(flags), sfxGroup, nil, sound)
		if result != C.MA_SUCCESS {
			return nil, fmt.Errorf("failed to load sound at %s, code %d", soundPath, result)
		}
		if looped {
			C.ma_sound_set_looping(sound, 1)
		} else {
			C.ma_sound_set_looping(sound, 0)
		}
		C.ma_sound_set_rolloff(sound, C.float(rolloff))
		C.ma_sound_set_min_distance(sound, 0.5)
		C.ma_sound_set_doppler_factor(sound, 0.0)
		C.ma_sound_set_pinned_listener_index(sound, 0)
		C.ma_sound_set_directional_attenuation_factor(sound, 0)
	}

	log.Printf("Sfx loaded at %v.\n", soundPath)
	return player, nil
}

func ListenerPosition() mgl32.Vec3 {
	vec := C.ma_engine_listener_get_position(&C.g_engine, 0)
	return mgl32.Vec3{
		float32(vec.x), float32(vec.y), float32(vec.z),
	}
}

func SetListenerPosition(newPos mgl32.Vec3) {
	C.ma_engine_listener_set_position(&C.g_engine, 0, C.float(newPos[0]), C.float(newPos[1]), C.float(newPos[2]))
}

func SetListenerOrientation(direction mgl32.Vec3) {
	C.ma_engine_listener_set_direction(&C.g_engine, 0, C.float(direction[0]), C.float(direction[1]), C.float(direction[2]))
}

func SetSfxVolume(volume float32) {
	C.ma_sound_group_set_volume(sfxGroup, C.float(volume))
}

func SfxVolume() float32 {
	return float32(C.ma_sound_group_get_volume(sfxGroup))
}

func SetMusicVolume(volume float32) {
	C.ma_sound_group_set_volume(musicGroup, C.float(volume))
}

func MusicVolume() float32 {
	return float32(C.ma_sound_group_get_volume(musicGroup))
}

func Update() {
	if nextSong != nil && (currentSong == nil || !currentSong.IsPlaying()) {
		if currentSong != nil {
			currentSong.Stop()
			currentSong.Teardown()
		}
		currentSong = nextSong
		if currentSong != nil {
			currentSong.Start()
		}
		nextSong = nil
	}
}

func QueueSong(path string, looping bool, fadeoutMs uint64) bool {
	if currentSong != nil && currentSong.IsPlaying() {
		C.ma_sound_stop_with_fade_in_milliseconds(&currentSong.sound, C.ma_uint64(fadeoutMs))
	}
	if nextSong != nil {
		nextSong.Teardown()
	}
	if path != "" {
		if currentSong == &songs[0] {
			nextSong = &songs[1]
		} else {
			nextSong = &songs[0]
		}
		cPath := C.CString(path)
		defer C.free(unsafe.Pointer(cPath))

		// Load sound data
		flags := C.ma_uint32(C.MA_SOUND_FLAG_STREAM | C.MA_SOUND_FLAG_NO_SPATIALIZATION)
		result := C.ma_sound_init_from_file(&C.g_engine, cPath, flags, musicGroup, nil, &nextSong.sound)
		if result != C.MA_SUCCESS {
			failure.LogErrWithLocation("failed to load song %v, code %v", path, result)
			goto abort
		}

		// Read the OGG metadata to find the looping points
		vorbError := C.int(0)
		vorb := C.stb_vorbis_open_filename(cPath, &vorbError, nil)
		if vorbError != 0 {
			failure.LogErrWithLocation("failed to read song metadata at %v, code %v", path, vorbError)
			goto abort
		}
		comments := C.stb_vorbis_get_comment(vorb)
		var loopStart, loopLength uint64
		for i := range comments.comment_list_length {
			comment := C.GoString((*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(comments.comment_list)) + uintptr(i))))
			var err error
			if strings.HasPrefix(comment, LoopStartTag) {
				loopStart, err = strconv.ParseUint(comment[len(LoopStartTag):], 10, 64)
				if err != nil {
					failure.LogErrWithLocation("could not parse loop start for %v", path)
				}
			}
			if strings.HasPrefix(comment, LoopLengthTag) {
				loopLength, err = strconv.ParseUint(comment[len(LoopLengthTag):], 10, 64)
				if err != nil {
					failure.LogErrWithLocation("could not parse loop length for %v", path)
				}
			}
		}
		if loopLength > 0 {
			result := C.ma_data_source_set_loop_point_in_pcm_frames(nextSong.sound.pDataSource, C.ma_uint64(loopStart), C.ma_uint64(loopStart)+C.ma_uint64(loopLength))
			if result != C.MA_SUCCESS {
				failure.LogErrWithLocation("failed to set loop point of song at %v, code %v", path, result)
				goto abort
			}
		}
		C.stb_vorbis_close(vorb)

		if looping {
			C.ma_sound_set_looping(&nextSong.sound, 1)
		} else {
			C.ma_sound_set_looping(&nextSong.sound, 0)
		}
	} else {
		nextSong = nil
	}
	return true
abort:
	nextSong = nil
	return false
}

func Teardown() {
	C.ma_engine_uninit(&C.g_engine)
	C.ma_device_uninit(device)
	C.ma_resource_manager_uninit(resourceManager)
	systemsPinner.Unpin()
}

func (voice *Voice) IsPlaying() bool {
	if voice == nil {
		return false
	}
	return C.ma_sound_is_playing(&voice.sound) == 1
}

// Returns the time that the sound has been playing in milliseconds.
func (voice *Voice) TimeMS() uint64 {
	if voice == nil {
		return 0
	}
	return uint64(C.ma_sound_get_time_in_milliseconds(&voice.sound))
}

// Returns the time that the sound has been playing in PCM frames.
func (voice *Voice) TimeFrames() uint64 {
	if voice == nil {
		return 0
	}
	return uint64(C.ma_sound_get_time_in_pcm_frames(&voice.sound))
}

func (voice *Voice) Start() {
	if voice == nil {
		return
	}
	result := C.ma_sound_start(&voice.sound)
	if result != C.MA_SUCCESS {
		failure.LogErrWithLocation("failed to start voice, code %v", result)
		return
	}
}

func (voice *Voice) Stop() {
	if voice == nil {
		return
	}
	C.ma_sound_stop(&voice.sound)
}

func (voice *Voice) SeekFrame(frame uint64) {
	if voice == nil {
		return
	}
	result := C.ma_sound_seek_to_pcm_frame(&voice.sound, C.ma_uint64(frame))
	if result != C.MA_SUCCESS {
		failure.LogErrWithLocation("seeking sound failed with code %v", result)
		return
	}
}

func (voice *Voice) SeekMS(time uint64) {
	if voice == nil {
		return
	}
	var soundLength C.ma_uint64
	result := C.ma_sound_get_length_in_pcm_frames(&voice.sound, &soundLength)
	if result != C.MA_SUCCESS {
		failure.LogErrWithLocation("failed to get sound length, code %v", result)
		return
	}
	pcmTime := time * uint64(SampleRate) / 1000
	if pcmTime >= uint64(soundLength) {
		pcmTime = uint64(soundLength) - 1
	}
	voice.SeekFrame(pcmTime)
}

func (voice *Voice) IsLooping() bool {
	if voice == nil {
		return false
	}
	return C.ma_sound_is_looping(&voice.sound) == 1
}

func (voice *Voice) Position() mgl32.Vec3 {
	if voice == nil {
		return mgl32.Vec3{}
	}
	maVec := C.ma_sound_get_position(&voice.sound)
	return mgl32.Vec3{
		float32(maVec.x),
		float32(maVec.y),
		float32(maVec.z),
	}
}

func (voice *Voice) SetPosition(newPos mgl32.Vec3) {
	if voice == nil {
		return
	}
	C.ma_sound_set_position(&voice.sound, C.float(newPos[0]), C.float(newPos[1]), C.float(newPos[2]))
}

func (voice *Voice) Teardown() {
	if voice == nil {
		return
	}
	C.ma_sound_uninit(&voice.sound)
}

// Attempts to play the sound using one of the available voices. Returns nil if there is not one available.
func (player *Player) Play() *Voice {
	var chosenVoice *Voice
	for v := range player.voices {
		voice := &player.voices[v]
		if !voice.IsPlaying() {
			chosenVoice = voice
			break
		} else if chosenVoice == nil || chosenVoice.TimeMS() < voice.TimeMS() {
			// Overwrite the oldest sound if they are all playing
			chosenVoice = voice
		}
	}
	if chosenVoice != nil {
		if chosenVoice.IsPlaying() {
			if player.noRetrigger {
				return nil
			} else {
				chosenVoice.Stop()
			}
		}

		chosenVoice.SeekFrame(0)
		C.ma_sound_set_spatialization_enabled(&chosenVoice.sound, 0)

		if result := C.ma_sound_start(&chosenVoice.sound); result != C.MA_SUCCESS {
			failure.LogErrWithLocation("failed to start sound, code %v", result)
			return nil
		}

		chosenVoice.playCount++
		return chosenVoice
	}
	return nil
}

func (player *Player) PlayAttenuated(position mgl32.Vec3) *Voice {
	listenerPos := ListenerPosition()
	var chosenVoice *Voice
	for v := range player.voices {
		voice := &player.voices[v]
		if !voice.IsPlaying() {
			chosenVoice = voice
			break
		}
		chosenVoiceDist := listenerPos.Sub(chosenVoice.Position()).LenSqr()
		voiceDist := listenerPos.Sub(voice.Position()).LenSqr()
		if chosenVoice == nil || chosenVoiceDist < voiceDist || (chosenVoiceDist == voiceDist && chosenVoice.TimeMS() < voice.TimeMS()) {
			// Overwrite the most distant, oldest sound if they are all playing
			chosenVoice = voice
		}
	}
	if chosenVoice != nil {
		if chosenVoice.IsPlaying() {
			if player.noRetrigger {
				return nil
			} else {
				chosenVoice.Stop()
			}
		}

		chosenVoice.SeekFrame(0)
		C.ma_sound_set_spatialization_enabled(&chosenVoice.sound, 1)
		chosenVoice.SetPosition(position)

		if result := C.ma_sound_start(&chosenVoice.sound); result != C.MA_SUCCESS {
			failure.LogErrWithLocation("failed to start sound, code %v", result)
			return nil
		}

		chosenVoice.playCount++
		return chosenVoice
	}
	return nil
}

func (player *Player) IsLooping() bool {
	if player == nil || len(player.voices) == 0 {
		return false
	}
	return player.voices[0].IsLooping()
}

func (player *Player) Teardown() {
	for _, voice := range player.voices {
		voice.Teardown()
	}
}

func (player *Player) StopAll() {
	for _, voice := range player.voices {
		voice.Stop()
	}
}
