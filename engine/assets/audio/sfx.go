package audio

import (
	"fmt"
	"log"
	"os"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
)

const ERROR_SFX_PATH = "assets/sounds/error.wav"

type sfxMetadata struct {
	Loop      bool
	Polyphony int
	Rolloff   *float32
}

func LoadSfx(soundPath string) (tdaudio.SoundId, error) {
	if !strings.HasSuffix(soundPath, ".wav") {
		return tdaudio.SoundId{}, fmt.Errorf("%v is not a wav file", soundPath)
	}

	// Look for metadata file
	metaPath := strings.TrimSuffix(soundPath, ".wav") + ".json"
	metadata, err := assets.LoadAndUnmarshalJSON[sfxMetadata](metaPath)
	if _, ok := err.(*os.PathError); err != nil && !ok {
		// The file is optional, so print errors that aren't 'file not found'.
		log.Printf("Could not parse metadata for %s; %s\n", soundPath, err)
	}

	polyphony := uint8(4)
	looped := false
	rolloff := float32(0.1)
	if metadata != nil {
		if metadata.Polyphony != 0 {
			polyphony = uint8(metadata.Polyphony)
		}
		looped = metadata.Loop
		if metadata.Rolloff != nil {
			rolloff = *metadata.Rolloff
		}
	}
	soundHandle := tdaudio.LoadSound(soundPath, polyphony, looped, rolloff)
	log.Printf("Sfx loaded at %v.\n", soundPath)
	return soundHandle, nil
}
