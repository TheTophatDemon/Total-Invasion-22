package assets

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/fzipp/bmfont"
)

type Font bmfont.Descriptor

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

func loadAngelcodeFont(path string) (*Font, error) {
	file, err := getFile(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dir, _ := filepath.Split(path)
	bmFont, err := bmfont.Read(file, fileSheets(dir))
	if err != nil {
		return nil, err
	}

	if len(bmFont.PageSheets) > 1 {
		return nil, fmt.Errorf("fonts with multiple pages are not supported")
	}

	font := Font(*bmFont.Descriptor)

	log.Printf("Angelcode font loaded at %v.\n", path)
	return &font, nil
}
