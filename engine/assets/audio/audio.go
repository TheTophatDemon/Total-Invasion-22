package audio

import (
	"sync/atomic"

	"github.com/ebitengine/oto/v3"
)

const (
	SAMPLE_RATE    = 44100
	CHANNEL_COUNT  = 2
	BUS_VOLUME_MAX = 100
)

var context *oto.Context

// Goes from 0-100
var sfxVolume, musicVolume atomic.Uint32

func Init() error {
	opt := oto.NewContextOptions{
		SampleRate:   SAMPLE_RATE,
		ChannelCount: CHANNEL_COUNT,
		Format:       oto.FormatFloat32LE,
	}
	var ready chan struct{}
	var err error
	context, ready, err = oto.NewContext(&opt)
	if err != nil {
		return err
	}
	<-ready

	sfxVolume.Store(1.0)
	musicVolume.Store(1.0)

	return nil
}

// Sets the overall volume for all sound effects. Takes a value from 0-1.
func SetSfxBusVolume(volume float32) {
	if volume > 1.0 {
		volume = 1.0
	} else if volume < 0.0 {
		volume = 0.0
	}
	sfxVolume.Store(uint32(volume * BUS_VOLUME_MAX))
}

// Sets the overall volume for all music. Takes a value from 0-1.
func SetMusicBusVolume(volume float32) {
	if volume > 1.0 {
		volume = 1.0
	} else if volume < 0.0 {
		volume = 0.0
	}
	musicVolume.Store(uint32(volume * BUS_VOLUME_MAX))
}
