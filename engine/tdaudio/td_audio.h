#ifndef TD_AUDIO_H
#define TD_AUDIO_H

#include <stdbool.h>
#include <stdint.h>

#define SAMPLE_RATE 44100
#define N_CHANNELS 2

typedef struct td_player_id {
    uint32_t id;
} td_player_id;

typedef struct td_voice_id {
    td_player_id player;
    uint32_t id, play_count;
} td_voice_id;

bool td_audio_init();
bool td_audio_voice_is_valid(td_voice_id voice);
td_player_id td_audio_load_sound(const char *path, uint8_t polyphony, bool looping, float rolloff);
td_voice_id td_audio_play_sound(td_player_id sound_id, float x, float y, float z, bool attenuated);
void td_audio_free_sounds(void);
bool td_audio_sound_is_looped(td_player_id sound);
bool td_audio_sound_is_playing(td_voice_id voice);
void td_audio_set_sound_position(td_voice_id voice, float x, float y, float z);
void td_audio_stop_sound(td_voice_id id);
void td_audio_seek_sound(td_voice_id voice, uint64_t time_ms);
uint64_t td_audio_get_sound_time(td_voice_id voice);
void td_audio_set_listener_orientation(float pos_x, float pos_y, float pos_z, float dir_x, float dir_y, float dir_z);
void td_audio_set_sfx_volume(float new_volume);
float td_audio_get_sfx_volume();
void td_audio_set_music_volume(float new_volume);
float td_audio_get_music_volume();
bool td_audio_queue_song(const char *path, bool looping, uint64_t fadeout_millis);
void td_audio_update();
void td_audio_teardown();

#endif