package audio

import "github.com/ebitengine/oto/v3"

const (
	SAMPLE_RATE   = 48000
	CHANNEL_COUNT = 2
)

var context *oto.Context

func Init() error {
	opt := oto.NewContextOptions{
		SampleRate:   SAMPLE_RATE,
		ChannelCount: CHANNEL_COUNT,
		Format:       oto.FormatSignedInt16LE,
	}
	var ready chan struct{}
	var err error
	context, ready, err = oto.NewContext(&opt)
	if err != nil {
		return err
	}
	<-ready

	return nil
}
