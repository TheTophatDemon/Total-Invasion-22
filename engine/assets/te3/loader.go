package te3

import (
	"log"

	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type TE3File struct {
	Meta struct {
		Editor  string `json:"editor"`
		Version string `json:"version"`
	} `json:"meta"`
	Ents     []Ent `json:"ents"`
	Tiles    Tiles `json:"tiles"`
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
