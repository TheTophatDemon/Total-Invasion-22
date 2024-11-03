package te3

import (
	"fmt"
	"log"

	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type TE3File struct {
	Ents     []Ent
	Tiles    Tiles
	filePath string
}

// Loads a Total Editor 3 map file into a data structure
func LoadTE3File(assetPath string) (*TE3File, error) {
	te3, err := assets.LoadAndUnmarshalJSON[TE3File](assetPath)
	te3.filePath = assetPath
	if err == nil {
		log.Println("Loaded TE3 file", assetPath)
	}
	return te3, err
}

func (te3 *TE3File) FilePath() string {
	return te3.filePath
}

// Returns the first entity in the map with the given key value pair in its properties.
func (te3 *TE3File) FindEntWithProperty(key, value string) (Ent, error) {
	for _, ent := range te3.Ents {
		if val, ok := ent.Properties[key]; ok && val == value {
			return ent, nil
		}
	}
	return Ent{}, fmt.Errorf("could not find entity with property tuple (%s, %s)", key, value)
}

// Returns an array of entities in the map with the given key value pair in their properties.
func (te3 *TE3File) FindEntsWithProperty(key string, values ...string) []Ent {
	out := make([]Ent, 0)
	for _, ent := range te3.Ents {
		for _, value := range values {
			if val, ok := ent.Properties[key]; ok && val == value {
				out = append(out, ent)
				break
			}
		}
	}
	return out
}
