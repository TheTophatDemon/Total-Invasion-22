#include "./td_audio.h"
#include <stdio.h>
#include <assert.h>

#define STB_VORBIS_HEADER_ONLY
#include "stb_vorbis.c"
#undef STB_VORBIS_HEADER_ONLY
#define MA_NO_MP3
#define MA_NO_FLAC
#define MA_NO_ENCODING
#define MINIAUDIO_IMPLEMENTATION
#include "miniaudio.h"

#define MAX($a, $b) (($a > $b) ? $a : $b)

#define ARRAY_PUSH($array, $type, $item) do {                                                        \
    ++$array.length;                                                                                 \
    if ($array.length >= $array.capacity) {                                                          \
        $array.capacity = MAX($array.capacity, 1) * 2;                                               \
        $array.items = ($type *)realloc($array.items, $array.capacity * sizeof($type));              \
    }                                                                                                \
    $array.items[$array.length-1] = ($item);                                                         \
} while(0)

#define LOG_ERR($message, ...) fprintf(stderr, "error occurred in " __FILE__ ":%s:%d " $message "\n", __func__, __LINE__, __VA_ARGS__)

#define SONG_QUEUE_MAX 2

typedef struct td_voice {
    ma_sound sound;
    uint32_t play_count;
} td_voice;

typedef struct td_player {
    uint8_t num_voices;
    td_voice *voices;
} td_player;

typedef struct td_player_list {
    uint32_t length;
    uint32_t capacity;
    td_player *items;
} td_player_list;

static ma_device g_device;
static ma_resource_manager g_resource_manager;
static ma_engine g_engine;
static ma_sound_group g_sfx_group, g_music_group;
static td_player_list g_players;
static ma_sound g_songs[SONG_QUEUE_MAX];
static ma_sound *g_current_song, *g_next_song;

static void data_callback(ma_device* pDevice, void* pOutput, const void* pInput, ma_uint32 frameCount)
{
    (void) pInput;

    ma_engine_read_pcm_frames(&g_engine, pOutput, frameCount, NULL);
}

void td_audio_free_sounds(void) {
    for (int s = 0; s < g_players.length; ++s) {
        td_player *player = &g_players.items[s];
        for (int v = 0; v < player->num_voices; ++v) {
            ma_sound_uninit(&player->voices[v].sound);
        }
        free(player->voices);
    }
    g_players.length = g_players.capacity = 0;
    free(g_players.items);
}

bool td_audio_voice_is_valid(td_voice_id voice) {
    if (voice.player.id == 0 || 
        voice.player.id >= g_players.length) {
        return false;
    }
    td_player *player = &g_players.items[voice.player.id];
    if (voice.id >= player->num_voices ||
        voice.play_count != player->voices[voice.id].play_count) {
        return false;
    }
    return true;
}

td_player_id td_audio_load_sound(const char *path, uint8_t polyphony, bool looping, float rolloff) {
    assert(path != NULL);
    assert(polyphony > 0);

    td_player player = (td_player){
        .num_voices = polyphony,
        .voices = (td_voice *) calloc(polyphony, sizeof(td_voice)),
    };
    td_player_id new_id = (td_player_id){.id = g_players.length};
    ARRAY_PUSH(g_players, td_player, player);

    // Set up voices
    int v = 0;
    for (v = 0; v < polyphony; ++v) {
        ma_sound *sound = &player.voices[v].sound;
        ma_uint32 flags = MA_SOUND_FLAG_DECODE;
        ma_result result = ma_sound_init_from_file(&g_engine, path, flags, &g_sfx_group, NULL, sound);
        if (result != MA_SUCCESS) {
            LOG_ERR("failed to load sound at %s, code %d", path, result);
            goto fail;
        }
        ma_sound_set_looping(sound, (ma_bool32) looping);
        ma_sound_set_rolloff(sound, rolloff);
        ma_sound_set_min_distance(sound, 0.5f);
        ma_sound_set_doppler_factor(sound, 0.0);
        ma_sound_set_pinned_listener_index(sound, 0);
        ma_sound_set_directional_attenuation_factor(sound, 0);
    }
    return new_id;
fail:
    for (v -= 1; v > 0; --v) {
        ma_sound_uninit(&player.voices[v].sound);
    }
    --g_players.length;
    free(player.voices);
    return (td_player_id){0};
}

bool td_audio_sound_is_looped(td_player_id player_id) {
    if (player_id.id >= g_players.length) return false;
    td_player *player = &g_players.items[player_id.id];
    if (player->num_voices == 0) return false;
    return ma_sound_is_looping(&player->voices[0].sound);
}

bool td_audio_init() {
    ma_result result = MA_SUCCESS;

    // Resource manager init
    {
        ma_resource_manager_config config = ma_resource_manager_config_init();
        config.decodedFormat = ma_format_f32;
        config.decodedChannels = N_CHANNELS;
        config.decodedSampleRate = SAMPLE_RATE;
        if ((result = ma_resource_manager_init(&config, &g_resource_manager)) != MA_SUCCESS) {
            LOG_ERR("failed to initialize mini audio resource manager, code %d", result);
            return false;
        }
    }

    // Device init
    {
        ma_device_config config = ma_device_config_init(ma_device_type_playback);
        config.playback.format = ma_format_f32;
        config.playback.channels = N_CHANNELS;
        config.sampleRate = SAMPLE_RATE;
        config.dataCallback = data_callback;
        if ((result = ma_device_init(NULL, &config, &g_device)) != MA_SUCCESS) {
            LOG_ERR("failed to initialize audio system, code %d", result);
            return false;
        }  
    }

    // Engine init
    {
        ma_engine_config config = ma_engine_config_init();
        config.pDevice = &g_device;
        config.pResourceManager = &g_resource_manager;

        if ((result = ma_engine_init(&config, &g_engine)) != MA_SUCCESS) {
            LOG_ERR("failed to initialize miniaudio engine, code %d", result);
            return false;
        }
    }

    // Initialize sound groups
    if ((result = ma_sound_group_init(&g_engine, 0, NULL, &g_sfx_group)) != MA_SUCCESS) {
        LOG_ERR("failed to initialize sfx group, code %d", result);
        return false;
    }
    if ((result = ma_sound_group_init(&g_engine, 0, NULL, &g_music_group)) != MA_SUCCESS) {
        LOG_ERR("failed to initialize music group, code %d", result);
        return false;
    }

    // Start the device
    if ((result = ma_device_start(&g_device)) != MA_SUCCESS) {
        LOG_ERR("failed to start playback device, code %d", result);
        return false;
    }

    // Initialize static variables
    g_players = (td_player_list) {0};
    g_current_song = g_next_song = NULL;
    for (int i = 0; i < SONG_QUEUE_MAX; ++i) {
        g_songs[i] = (ma_sound) {0};
    }

    return true;
}

/// Attempts to play the sound using one of the available voices. Returns a zeroed voice ID if sound was not played.
/// `x`, `y`, and `z` are the spatial coordinates of the sound in world space. If `attenuated` is false, then those parameters don't do anything.
td_voice_id td_audio_play_sound(td_player_id player_id, float x, float y, float z, bool attenuated) {
    assert(player_id.id < g_players.length);

    td_player *player = &g_players.items[player_id.id];

    td_voice *chosen_voice = NULL;
    int chosen_voice_id = -1;
    ma_vec3f listener_pos = ma_engine_listener_get_position(&g_engine, 0);

    for (int v = 0; v < player->num_voices; ++v) {
        td_voice *voice = &player->voices[v];
        if (!ma_sound_is_playing(&voice->sound)) {
            chosen_voice = voice;
            chosen_voice_id = v;
            break;
        }
        float chosen_voice_dist_from_listener = ma_vec3f_len2(ma_vec3f_sub(listener_pos, ma_sound_get_position(&chosen_voice->sound)));
        float voice_dist_from_listener = ma_vec3f_len2(ma_vec3f_sub(listener_pos, ma_sound_get_position(&voice->sound)));
        if (chosen_voice == NULL || 
            chosen_voice_dist_from_listener < voice_dist_from_listener ||
            (chosen_voice_dist_from_listener == voice_dist_from_listener && ma_sound_get_time_in_milliseconds(&chosen_voice->sound) < ma_sound_get_time_in_milliseconds(&voice->sound))
        ) {
            // Overwrite the oldest, most distant sound.
            chosen_voice = voice;
            chosen_voice_id = v;
        }
    }

    if (chosen_voice != NULL && chosen_voice_id > -1) {
        ma_result result;
        
        if (ma_sound_is_playing(&chosen_voice->sound)) {
            ma_sound_stop(&chosen_voice->sound);
        }

        result = ma_sound_seek_to_pcm_frame(&chosen_voice->sound, 0);
        if (result != MA_SUCCESS) LOG_ERR("failed to seek sound with id %d, code %d", player_id.id, result);

        ma_sound_set_spatialization_enabled(&chosen_voice->sound, attenuated);
        if (attenuated) {
            ma_sound_set_position(&chosen_voice->sound, x, y, z);
        }

        result = ma_sound_start(&chosen_voice->sound);
        if (result != MA_SUCCESS) LOG_ERR("failed to start sound with id %d, code %d", player_id.id, result);
        
        ++chosen_voice->play_count;

        return (td_voice_id) {
            .player = player_id,
            .id = (uint32_t)chosen_voice_id,
            .play_count = chosen_voice->play_count,
        };
    }
    return (td_voice_id) {0};
}

bool td_audio_sound_is_playing(td_voice_id voice) {
    if (!td_audio_voice_is_valid(voice)) return false;
    td_player *player = &g_players.items[voice.player.id];
    return (bool)ma_sound_is_playing(&player->voices[voice.id].sound);
}

void td_audio_set_sound_position(td_voice_id voice, float x, float y, float z) {
    if (!td_audio_voice_is_valid(voice)) return;
    td_player *player = &g_players.items[voice.player.id];
    ma_sound *ma_player = &player->voices[voice.id].sound;
    if (!ma_sound_is_spatialization_enabled(ma_player)) return;
    ma_sound_set_position(ma_player, x, y, z);
}

void td_audio_stop_sound(td_voice_id voice) {
    if (!td_audio_voice_is_valid(voice)) return;
    td_player *player = &g_players.items[voice.player.id];
    ma_sound_stop(&player->voices[voice.id].sound);
}

void td_audio_seek_sound(td_voice_id voice, uint64_t time_ms) {
    if (!td_audio_voice_is_valid(voice)) return;
    td_player *player = &g_players.items[voice.player.id];
    ma_sound *ma_player = &player->voices[voice.id].sound;
    ma_uint64 sound_length;
    ma_result result = ma_sound_get_length_in_pcm_frames(ma_player, &sound_length);
    if (result != MA_SUCCESS) LOG_ERR("failed to get sound length with id %d, code %d", voice.player.id, result);
    ma_uint64 time_pcm = (time_ms * ma_engine_get_sample_rate(&g_engine)) / 1000;
    if (time_pcm >= sound_length) time_pcm = sound_length - 1;
    result = ma_sound_seek_to_pcm_frame(ma_player, time_pcm);
    if (result != MA_SUCCESS) LOG_ERR("failed to seek sound with id %d, code %d", voice.player.id, result);
}

uint64_t td_audio_get_sound_time(td_voice_id voice) {
    if (!td_audio_voice_is_valid(voice)) return 0;
    td_player *player = &g_players.items[voice.player.id];
    ma_sound *ma_player = &player->voices[voice.id].sound;
    return (uint64_t) ma_sound_get_time_in_milliseconds(ma_player);
}

void td_audio_set_listener_orientation(float pos_x, float pos_y, float pos_z, float dir_x, float dir_y, float dir_z) {
    ma_engine_listener_set_position(&g_engine, 0, pos_x, pos_y, pos_z);
    ma_engine_listener_set_direction(&g_engine, 0, dir_x, dir_y, dir_z);
}

void td_audio_set_sfx_volume(float new_volume) {
    ma_sound_group_set_volume(&g_sfx_group, new_volume);
}

float td_audio_get_sfx_volume() {
    return ma_sound_group_get_volume(&g_sfx_group);
}

void td_audio_set_music_volume(float new_volume) {
    ma_sound_group_set_volume(&g_music_group, new_volume);
}

float td_audio_get_music_volume() {
    return ma_sound_group_get_volume(&g_music_group);
}

bool td_audio_queue_song(const char *path, bool looping, uint64_t fadeout_millis) {
    if (g_current_song != NULL && ma_sound_is_playing(g_current_song)) {
        ma_sound_stop_with_fade_in_milliseconds(g_current_song, fadeout_millis);
    }
    if (g_next_song != NULL) {
        ma_sound_uninit(g_next_song);
    }
    if (path != NULL) {
        g_next_song = (g_current_song == &g_songs[0]) ? &g_songs[1] : &g_songs[0];
        ma_uint32 flags = MA_SOUND_FLAG_STREAM | MA_SOUND_FLAG_NO_SPATIALIZATION;
        ma_result result = ma_sound_init_from_file(&g_engine, path, flags, &g_music_group, NULL, g_next_song);
        if (result != MA_SUCCESS) {
            LOG_ERR("failed to load song at %s, code %d", path, result);
            g_next_song = NULL;
            return false;
        }
        ma_sound_set_looping(g_next_song, looping);
    } else {
        g_next_song = NULL;
    }
    return true;
}

void td_audio_update() {
    ma_result result;
    // Swap out the music track with the next one when ready.
    if (g_next_song != NULL && (g_current_song == NULL || !ma_sound_is_playing(g_current_song))) {
        if (g_current_song != NULL) ma_sound_uninit(g_current_song);
        g_current_song = g_next_song;
        if ((result = ma_sound_start(g_current_song)) != MA_SUCCESS) {
            LOG_ERR("failed to start new song, code %d", result);
            ma_sound_uninit(g_current_song);
            g_current_song = NULL;
        }
        g_next_song = NULL;
    }
}

void td_audio_teardown() {
    td_audio_free_sounds();
    ma_engine_uninit(&g_engine);
    ma_device_uninit(&g_device);
    ma_resource_manager_uninit(&g_resource_manager);
}