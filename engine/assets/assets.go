package assets

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Retrieves the asset's file from one of the available asset packs
func GetFile(assetPath string) (*os.File, error) {
	return os.Open(strings.ReplaceAll(assetPath, "\\", "/"))
}

// Returns an iterator over all files in the available asset packs under the given directory recursively.
// Can filter by file extension using the `withExtension` parameter, or leave it blank.
func GetFileNamesFromDir(directory, withExtension string) ([]string, error) {
	result := make([]string, 0, 32)
	searchQueue := []string{directory}
	var allErrors error
	for len(searchQueue) > 0 {
		searchPath := searchQueue[len(searchQueue)-1]
		searchQueue = searchQueue[:len(searchQueue)-1]

		info, err := os.Stat(searchPath)
		if err != nil {
			allErrors = errors.Join(allErrors, err)
			continue
		}

		if info.IsDir() {
			entries, err := os.ReadDir(searchPath)
			if err != nil {
				allErrors = errors.Join(allErrors, err)
				continue
			}
			for _, entry := range entries {
				searchQueue = append(searchQueue, filepath.Join(searchPath, entry.Name()))
			}
		} else if strings.HasSuffix(searchPath, withExtension) {
			result = append(result, searchPath)
		}
	}
	return result, allErrors
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
