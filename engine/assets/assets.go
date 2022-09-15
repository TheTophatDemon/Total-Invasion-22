package assets

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

var textures map[string]Texture
var meshes   map[string]*Mesh

func init() {
	textures = make(map[string]Texture)
	meshes   = make(map[string]*Mesh)
}

//Retrieves the asset's file from one of the available asset packs
func getFile(assetPath string) (*os.File, error) {
	return os.Open(assetPath)
}

func GetTexture(assetPath string) Texture {
	texture, ok := textures[assetPath]
	if !ok {
		textures[assetPath] = loadTexture(assetPath)
		return textures[assetPath]
	}
	return texture
}

func GetMesh(assetPath string) (*Mesh, error) {
	var err error
	
	mesh, ok := meshes[assetPath]
	if !ok {
		switch {
		case strings.HasSuffix(assetPath, ".obj"):
			mesh, err = loadOBJMesh(assetPath)
		default:
			err = fmt.Errorf("Unsupported model file type.")
		}
		if err != nil {
			return nil, nil
		}
		meshes[assetPath] = mesh
	}
	return mesh, nil
}

//Loads a JSON file from the given asset-path (relative to assets folder) and returns the json.Unmarshal result as type T.
func LoadAndUnmarshalJSON[T any](assetPath string) (*T, error) {
	file, err := getFile(assetPath)
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