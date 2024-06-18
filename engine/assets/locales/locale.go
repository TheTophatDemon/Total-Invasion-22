package locales

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/assets"
)

const (
	ENGLISH = "en"
	RUSSIAN = "ru"
)

type (
	Translation map[string]string
)

func (trans *Translation) UnmarshalJSON(b []byte) error {
	// Get JSON data as a map
	var jData map[string]any
	if err := json.Unmarshal(b, &jData); err != nil {
		return err
	}

	*trans = make(Translation)

	var err error = nil
	for key, value := range jData {
		switch val := value.(type) {
		case string:
			(*trans)[key] = val
		case []string:
			(*trans)[key] = strings.Join(val, "\n")
		default:
			if err == nil {
				err = fmt.Errorf("locale values should be string or array of strings")
			}
		}
	}

	return err
}

func LoadTranslation(assetPath string) (*Translation, error) {
	trans, err := assets.LoadAndUnmarshalJSON[Translation](assetPath)
	if err != nil {
		return nil, err
	}

	log.Printf("Translation loaded at %v.\n", assetPath)
	return trans, nil
}
