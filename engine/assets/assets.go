package assets

import (
	_ "embed"
	"encoding/json"
	"io"
	"os"
)

// Retrieves the asset's file from one of the available asset packs
func GetFile(assetPath string) (*os.File, error) {
	return os.Open(assetPath)
}

// Loads a JSON file from the given asset-path (relative to assets folder) and returns the json.Unmarshal result as type T.
func LoadAndUnmarshalJSON[T any](assetPath string) (*T, error) {
	file, err := GetFile(assetPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	t := new(T)
	err = json.Unmarshal(fileBytes, t)

	return t, err
}
