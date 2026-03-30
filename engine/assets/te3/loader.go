package te3

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type EditorCamera struct {
	EulerAngles mgl32.Vec3 `json:"eulerAngles"`
	Position    mgl32.Vec3 `json:"position"`
}

type TE3File struct {
	EditorCamera EditorCamera `json:"editorCamera"`
	Meta         struct {
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
	if err != nil {
		//TODO: Handle this error gracefully by returning to the title screen.
		return nil, err
	}
	te3.filePath = assetPath
	log.Println("Loaded TE3 file", assetPath)
	return te3, err
}

func (te3 *TE3File) FilePath() string {
	return te3.filePath
}
