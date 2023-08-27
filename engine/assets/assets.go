package assets

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
)

// This map caches the textures loaded from the filesystem by their paths.
var textures map[string]*Texture

// This caches the meshes loaded from the filesystem by their paths.
var meshes map[string]*Mesh

// Cache of bitmap fonts, indexed by .fnt file path
var fonts map[string]*Font

func init() {
	textures = make(map[string]*Texture)
	meshes = make(map[string]*Mesh)
	fonts = make(map[string]*Font)
}

var (
	// A quad on the XY plane centered at (0,0) with a width and height of 2.
	QuadMesh *Mesh
)

// Initialize built-in assets
func Init() {
	shaders.Init()

	QuadMesh = CreateMesh(Vertices{
		Pos: []mgl32.Vec3{
			{-1.0, -1.0, 0.0},
			{1.0, -1.0, 0.0},
			{-1.0, 1.0, 0.0},
			{1.0, 1.0, 0.0},
		},
		TexCoord: []mgl32.Vec2{
			{1.0, 0.0},
			{0.0, 0.0},
			{1.0, 1.0},
			{0.0, 1.0},
		},
		Normal: []mgl32.Vec3{
			{0.0, 0.0, 1.0},
			{0.0, 0.0, 1.0},
			{0.0, 0.0, 1.0},
			{0.0, 0.0, 1.0},
		},
		Color: nil,
	}, []uint32{
		1, 2, 0, 1, 3, 2,
	})
}

func FreeBuiltInAssets() {
	shaders.Free()
}

// Retrieves the asset's file from one of the available asset packs
func getFile(assetPath string) (*os.File, error) {
	return os.Open(assetPath)
}

func GetTexture(assetPath string) *Texture {
	texture, ok := textures[assetPath]
	if !ok {
		textures[assetPath] = loadTexture(assetPath)
		return textures[assetPath]
	}
	return texture
}

// Releases memory for all cached textures
func FreeTextures() {
	for p := range textures {
		textures[p].Free()
		delete(textures, p)
	}
}

// Releases memory for all cached meshes
func FreeMeshes() {
	for p := range meshes {
		meshes[p].Free()
		delete(meshes, p)
	}
}

// Releases memory for all cached fonts
func FreeFonts() {
	clear(fonts)
}

func FreeAll() {
	FreeTextures()
	FreeMeshes()
	FreeFonts()
	FreeBuiltInAssets()
}

// Takes ownership of an already loaded mesh. Will dispose of its resources along with the other meshes.
func TakeMesh(assetPath string, mesh *Mesh) {
	meshes[assetPath] = mesh
}

func GetMesh(assetPath string) (*Mesh, error) {
	var err error

	mesh, ok := meshes[assetPath]
	if !ok {
		switch {
		case strings.HasSuffix(assetPath, ".obj"):
			mesh, err = loadOBJMesh(assetPath)
		default:
			err = fmt.Errorf("unsupported model file type")
		}
		if err != nil {
			return nil, err
		}
		meshes[assetPath] = mesh
	}
	return mesh, nil
}

// Retrieves a font from the game assets, loading it if it doesn't already exist.
func GetFont(assetPath string) (*Font, error) {
	var err error

	font, ok := fonts[assetPath]
	if !ok {
		// Check file extension
		if !strings.HasSuffix(assetPath, ".fnt") {
			return nil, fmt.Errorf("unsupported font file type")
		}

		font, err = loadAngelcodeFont(assetPath)
		if err != nil {
			return nil, err
		}

		fonts[assetPath] = font
	}
	return font, nil
}

// Loads a JSON file from the given asset-path (relative to assets folder) and returns the json.Unmarshal result as type T.
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
