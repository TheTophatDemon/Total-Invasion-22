package locales

import (
	"log"

	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type Translation map[string]string

func LoadTranslation(assetPath string) (*Translation, error) {
	trans, err := assets.LoadAndUnmarshalTOML[Translation](assetPath)
	if err != nil {
		return nil, err
	}

	log.Printf("Translation loaded at %v.\n", assetPath)
	return trans, nil
}
