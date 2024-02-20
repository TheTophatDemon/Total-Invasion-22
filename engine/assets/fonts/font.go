package fonts

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fzipp/bmfont"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type Font struct {
	bmfont.Descriptor
	texturePath string
}

func fileSheets(directory string) bmfont.SheetReaderFunc {
	return func(filename string) (io.ReadCloser, error) {
		path := filepath.Join(directory, filename)
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
}

func LoadAngelcodeFont(assetPath string) (*Font, error) {
	file, err := assets.GetFile(assetPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dir, _ := filepath.Split(assetPath)
	bmFont, err := bmfont.Read(file, fileSheets(dir))
	if err != nil {
		return nil, err
	}

	if len(bmFont.PageSheets) > 1 {
		return nil, fmt.Errorf("fonts with multiple pages are not supported")
	}

	font := &Font{
		Descriptor: *bmFont.Descriptor,
	}

	// Set the texture to the font's page
	font.texturePath = path.Join(path.Dir(assetPath), font.Pages[0].File)
	if !strings.HasSuffix(font.texturePath, ".png") {
		return nil, fmt.Errorf("font has invalid texture file type")
	}

	log.Printf("Angelcode font loaded at %v.\n", assetPath)
	return font, nil
}

func (f *Font) TexturePath() string {
	return f.texturePath
}
