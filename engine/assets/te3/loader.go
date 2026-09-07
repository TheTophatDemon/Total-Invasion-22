package te3

import (
	"encoding/json/v2"
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type EditorCamera struct {
	EulerAngles mgl32.Vec3 `json:"eulerAngles"`
	Position    mgl32.Vec3 `json:"position"`
}

type TE3File[EntType any] struct {
	EditorCamera EditorCamera `json:"editorCamera"`
	Meta         struct {
		Editor  string `json:"editor"`
		Version string `json:"version"`
	} `json:"meta"`
	Ents     []EntType `json:"ents"`
	Tiles    Tiles     `json:"tiles"`
	filePath string
}

// Returns options that should be used when parsing JSON on a .te3 file.
func ParseOptions() json.Options {
	return json.JoinOptions(
		json.DefaultOptionsV2(),
		json.RejectUnknownMembers(true),
	)
}

// Loads a Total Editor 3 map file into a data structure
func LoadTE3File[EntType any](assetPath string) (*TE3File[EntType], error) {
	te3, err := assets.LoadAndUnmarshalJSON[TE3File[EntType]](
		assetPath, ParseOptions(),
	)
	if err != nil {
		return nil, err
	}
	te3.filePath = assetPath
	log.Println("Loaded TE3 file", assetPath)
	return te3, err
}

func (te3 *TE3File[EntType]) FilePath() string {
	return te3.filePath
}
