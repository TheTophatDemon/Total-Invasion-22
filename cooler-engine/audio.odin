package audio

import "core:strconv"
import "core:strings"
import "core:math/linalg"
import "core:log"
import "base:runtime"

import ma "vendor:miniaudio"
import "vendor:stb/vorbis"

LOOP_START_TAG :: "LOOPSTART="
LOOP_LENGTH_TAG :: "LOOPLENGTH="

SAMPLE_RATE :: 44100
N_CHANNELS :: 2

Player_Id :: distinct u32
Voice_Id :: struct {
	player: Player_Id,
	id, play_count: u32,
}

Voice :: struct {
	sound: ma.sound,
	play_count: u32,
}

Player :: struct {
	voices: [dynamic]Voice,
	no_retrigger: bool,
}

g_device: ma.device
g_resource_manager: ma.resource_manager
g_engine: ma.engine
g_sfx_group, g_music_group: ma.sound_group
g_players: [dynamic]Player
g_songs: [2]ma.sound
g_current_song, g_next_song: ^ma.sound

data_callback :: proc "c" (device: ^ma.device, output: rawptr, input: rawptr, frame_count: u32) {
	ma.engine_read_pcm_frames(&g_engine, output, u64(frame_count), nil)
}

from_ma_vec3f :: proc "c" (ma_vec: ma.vec3f) -> [3]f32 {
	return {ma_vec.x, ma_vec.y, ma_vec.z}
}

@(export)
free_sounds :: proc "c" () {
	context = runtime.default_context()
	for &player in g_players {
		for &voice in player.voices {
			ma.sound_uninit(&voice.sound)
		}
		delete(player.voices)
	}
	clear(&g_players)
}

@(export)
voice_is_valid :: proc "c" (voice_id: Voice_Id) -> bool {
	context = runtime.default_context()
	if voice_id.player == 0 || int(voice_id.player) >= len(g_players) {
		return false
	}
	player := &g_players[voice_id.player]
	if int(voice_id.id) >= len(player.voices) || voice_id.play_count != player.voices[voice_id.id].play_count {
		return false
	}
	return true
}

@(export)
load_sound :: proc "c" (path: cstring, polyphony: u8, looping: bool, rolloff: f32, no_retrigger: bool) -> Player_Id {
	context = runtime.default_context()
	assert(path != nil)
	assert(polyphony > 0)

	player := Player{
		no_retrigger = no_retrigger,
		voices = make([dynamic]Voice, polyphony),
	}
	new_id := Player_Id(len(g_players))
	append(&g_players, player)

	// Set up vocies
	for v in 0..<len(player.voices) {
		sound := &player.voices[v].sound
		result := ma.sound_init_from_file(&g_engine, path, {.DECODE}, &g_sfx_group, nil, sound)
		if result != .SUCCESS {
			log.errorf("failed to load sound at %v, code %v", path, result)
			return 0
		}
		ma.sound_set_looping(sound, b32(looping))
		ma.sound_set_rolloff(sound, rolloff)
		ma.sound_set_min_distance(sound, 0.5)
		ma.sound_set_doppler_factor(sound, 0.0)
		ma.sound_set_pinned_listener_index(sound, 0)
		ma.sound_set_directional_attenuation_factor(sound, 0)
	}

	return new_id
}

@(export)
sound_is_looped :: proc "c" (player_id: Player_Id) -> bool {
	if int(player_id) >= len(g_players) do return false
	player := g_players[player_id]
	if len(player.voices) == 0 do return false
	return bool(ma.sound_is_looping(&player.voices[0].sound))
}

@(export)
audio_init :: proc "c" () -> bool {
	context = runtime.default_context()
	// Resource manager init
	{
		config := ma.resource_manager_config_init()
		config.decodedFormat = .f32
		config.decodedChannels = N_CHANNELS
		config.decodedSampleRate = SAMPLE_RATE
		if result := ma.resource_manager_init(&config, &g_resource_manager); result != .SUCCESS {
			log.errorf("failed to initialize miniaudio resource manager, code %v", result)
			return false
		}
	}

	// Device init
	{
        config := ma.device_config_init(.playback)
        config.playback.format = .f32
        config.playback.channels = N_CHANNELS
        config.sampleRate = SAMPLE_RATE
        config.dataCallback = data_callback
        if result := ma.device_init(nil, &config, &g_device); result != .SUCCESS {
            log.errorf("failed to initialize audio system, code %v", result)
            return false
        }
    }

	// Engine init
    {
        config := ma.engine_config_init();
        config.pDevice = &g_device;
        config.pResourceManager = &g_resource_manager;

        if result := ma.engine_init(&config, &g_engine); result != .SUCCESS {
            log.errorf("failed to initialize miniaudio engine, code %v", result);
            return false
        }
    }

    // Initialize sound groups
    if result := ma.sound_group_init(&g_engine, {}, nil, &g_sfx_group); result != .SUCCESS {
        log.errorf("failed to initialize sfx group, code %v", result);
        return false
    }
    if result := ma.sound_group_init(&g_engine, {}, nil, &g_music_group); result != .SUCCESS {
        log.errorf("failed to initialize music group, code %v", result);
        return false
    }

    // Start the device
    if result := ma.device_start(&g_device); result != .SUCCESS {
        log.errorf("failed to start playback device, code %v", result);
        return false
    }

	return true
}

// Attempts to play the sound using one of the available voices. Returns a zeroed voice ID if sound was not played.
// `x`, `y`, and `z` are the spatial coordinates of the sound in world space. If `attenuated` is false, then those parameters don't do anything.
@(export)
play_sound :: proc "c" (player_id: Player_Id, x, y, z: f32, attenuated: bool) -> Voice_Id {
	context = runtime.default_context()
	assert(int(player_id) < len(g_players))
	player := &g_players[player_id]
	chosen_voice: ^Voice = nil
	chosen_voice_id := -1
	listener_pos := from_ma_vec3f(ma.engine_listener_get_position(&g_engine, 0))
	for &voice, v in player.voices {
		if !ma.sound_is_playing(&voice.sound) {
			chosen_voice = &voice
			chosen_voice_id = v
			break
		}
		
		chosen_voice_dist_from_listener: f32
		if chosen_voice != nil {
			chosen_voice_dist_from_listener = linalg.length2(listener_pos - from_ma_vec3f(ma.sound_get_position(&chosen_voice.sound)))
		}
		voice_dist_from_listener := linalg.length2(listener_pos - from_ma_vec3f(ma.sound_get_position(&voice.sound)))
		if chosen_voice == nil || 
			chosen_voice_dist_from_listener < voice_dist_from_listener || 
			(chosen_voice_dist_from_listener == voice_dist_from_listener && ma.sound_get_time_in_milliseconds(&chosen_voice.sound) < ma.sound_get_time_in_milliseconds(&voice.sound))
		{
			// Overwrite the oldest, most distant sound.
			chosen_voice = &voice
			chosen_voice_id = v	
		}
	}

	if chosen_voice != nil && chosen_voice_id > -1 {
		if ma.sound_is_playing(&chosen_voice.sound) {
			if player.no_retrigger {
				return {}
			} else {
				ma.sound_stop(&chosen_voice.sound)
			}
		}

		if result := ma.sound_seek_to_pcm_frame(&chosen_voice.sound, 0); result != .SUCCESS {
			log.errorf("failed to seek sound with id %v, code %v", player_id, result)
		}

		ma.sound_set_spatialization_enabled(&chosen_voice.sound, b32(attenuated))
		if (attenuated) {
			ma.sound_set_position(&chosen_voice.sound, x, y, z)
		}

		if result := ma.sound_start(&chosen_voice.sound); result != .SUCCESS {
			log.errorf("failed to start sound with id %v, code %v", player_id, result)
		}

		chosen_voice.play_count += 1

		return Voice_Id {
			player = player_id,
			id = u32(chosen_voice_id),
			play_count = chosen_voice.play_count,
		}
	}

	return {}
}

@(export)
sound_is_playing :: proc "c" (voice: Voice_Id) -> bool {
    if !voice_is_valid(voice) do return false
    player := &g_players[voice.player]
    return bool(ma.sound_is_playing(&player.voices[voice.id].sound))
}

@(export)
sound_set_position :: proc "c" (voice: Voice_Id, x, y, z: f32) {
	if !voice_is_valid(voice) do return
	player := &g_players[voice.player]
	ma_player := &player.voices[voice.id].sound
	if !ma.sound_is_spatialization_enabled(ma_player) do return
	ma.sound_set_position(ma_player, x, y, z)
}

@(export)
stop_sound :: proc "c" (voice: Voice_Id) {
	if !voice_is_valid(voice) do return
	player := &g_players[voice.player]
	ma.sound_stop(&player.voices[voice.id].sound)
}

@(export)
stop_all_sounds :: proc "c" () {
	for &player in g_players {
		for &voice in player.voices {
			ma.sound_stop(&voice.sound)
		}
	}
}

@(export)
seek_sound :: proc "c" (voice: Voice_Id, time_ms: u64) {
	context = runtime.default_context()
	if !voice_is_valid(voice) do return
	player := &g_players[voice.player]
	ma_player := &player.voices[voice.id].sound
	sound_length: u64
	if result := ma.sound_get_length_in_pcm_frames(ma_player, &sound_length); result != .SUCCESS {
		log.errorf("failed to get sound length with id %v, code %v", voice.player, result)
	}
	time_pcm := (time_ms * u64(ma.engine_get_sample_rate(&g_engine))) / 1000
	if time_pcm >= sound_length do time_pcm = sound_length - 1
	if result := ma.sound_seek_to_pcm_frame(ma_player, time_pcm); result != .SUCCESS {
		log.errorf("failed to seek sound with id %v, code %v", voice.player, result)
	}
}

@(export)
get_sound_time :: proc "c" (voice: Voice_Id) -> u64 {
	if !voice_is_valid(voice) do return 0
	player := &g_players[voice.player]
	ma_player := &player.voices[voice.id].sound
	return ma.sound_get_time_in_milliseconds(ma_player)
}

@(export)
set_listener_orientation :: proc "c" (pos_x, pos_y, pos_z, dir_x, dir_y, dir_z: f32) {
	ma.engine_listener_set_position(&g_engine, 0, pos_x, pos_y, pos_z)
	ma.engine_listener_set_direction(&g_engine, 0, dir_x, dir_y, dir_z)
}

@(export)
set_sfx_volume :: proc "c" (new_volume: f32) {
	ma.sound_group_set_volume(&g_sfx_group, new_volume)
}

@(export)
get_sfx_volume :: proc "c" () -> f32 {
	return ma.sound_group_get_volume(&g_sfx_group)
}

@(export)
set_music_volume :: proc "c" (new_volume: f32) {
	ma.sound_group_set_volume(&g_music_group, new_volume)
}

@(export)
get_music_volume :: proc "c" () -> f32 {
	return ma.sound_group_get_volume(&g_music_group)
}

@(export)
queue_song :: proc "c" (path: cstring, looping: bool, fadeout_millis: u64) -> bool {
	context = runtime.default_context()
	happy_path: {
		if g_current_song != nil && ma.sound_is_playing(g_current_song) {
			ma.sound_stop_with_fade_in_milliseconds(g_current_song, fadeout_millis)
		}
		if g_next_song != nil {
			ma.sound_uninit(g_next_song)
		}
		if path != nil && path != "" {
			g_next_song = (g_current_song == &g_songs[0]) ? &g_songs[1] : &g_songs[0]
			result := ma.sound_init_from_file(
				&g_engine, 
				path, 
				{ .STREAM, .NO_SPATIALIZATION }, 
				&g_music_group, 
				nil, 
				g_next_song,
			)
			if result != .SUCCESS {
				log.errorf("failed to load song at %v, code %v", path, result)
				break happy_path
			}

			// Read the OGG metadata to find the looping points
			vorb_error: vorbis.Error
			vorb := vorbis.open_filename(path, &vorb_error, nil)
			if vorb_error != .none {
				log.errorf("failed to read song metadata at %v, code %v", path, vorb_error)
				break happy_path
			}
			comments := vorbis.get_comment(vorb)
			
			loop_start, loop_length: u64
			for i in 0..<comments.comment_list_length {
				if comments.comment_list[i] == nil do continue
				comment := string(comments.comment_list[i])
				ok: bool
				if strings.has_prefix(comment, LOOP_START_TAG) {
					loop_start, ok = strconv.parse_u64(comment[len(LOOP_START_TAG):])
					if !ok {
						log.warnf("could not parse loop start for song %v", path)
					}
				}
				if strings.has_prefix(comment, LOOP_LENGTH_TAG) {
					loop_length, ok = strconv.parse_u64(comment[len(LOOP_LENGTH_TAG):])
					if !ok {
						log.warnf("could not parse loop length for song %v", path)
					}
				}
				if loop_length > 0 {
					if result := ma.data_source_set_loop_point_in_pcm_frames(g_next_song.pDataSource, loop_start, loop_start + loop_length); result != .SUCCESS {
						log.errorf("failed to set loop point of song at %v, code %v", path, result)
						break happy_path
					}
				}
			}
			vorbis.close(vorb)

			ma.sound_set_looping(g_next_song, b32(looping))
		} else {
			g_next_song = nil
		}
		return true
	}
	g_next_song = nil
	return false
}

@(export)
audio_update :: proc "c" () {
	context = runtime.default_context()
	// Swap out the music track with the next one when ready.
	if g_next_song != nil && (g_current_song == nil || !ma.sound_is_playing(g_current_song)) {
		if g_current_song != nil do ma.sound_uninit(g_current_song)
		g_current_song = g_next_song
		if result := ma.sound_start(g_current_song); result != .SUCCESS {
			log.errorf("failed to start new song, code %v", result)
			ma.sound_uninit(g_current_song)
			g_current_song = nil
		}
		g_next_song = nil
	}
} 

@(export)
audio_teardown :: proc "c" () {
	free_sounds()
	ma.engine_uninit(&g_engine)
	ma.device_uninit(&g_device)
	ma.resource_manager_uninit(&g_resource_manager)
}