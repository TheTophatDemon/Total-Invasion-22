package audio

import (
	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/wav"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

const SFX_MAX_PLAYERS = 8

type Sfx struct {
	decoder wav.Decoder
	players [SFX_MAX_PLAYERS]oto.Player
}

func LoadSfx(assetPath string) (*Sfx, error) {
	file, err := assets.GetFile(assetPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sfx := &Sfx{
		decoder: *wav.NewDecoder(file),
	}
	for i := range sfx.players {
		sfx.players[i] = *context.NewPlayer(sfx.decoder.PCMBuffer())
	}
	return sfx, nil
}
