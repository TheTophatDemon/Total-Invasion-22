package cache

import (
	"fmt"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/fonts"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/shaders"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
)

// This map caches the loadedTextures loaded from the filesystem by their paths.
var loadedTextures map[string]*textures.Texture

// This caches the meshes loaded from the filesystem by their paths.
var loadedMeshes map[string]*geom.Mesh

// Cache of bitmap fonts, indexed by .fnt file path
var loadedFonts map[string]*fonts.Font

func init() {
	loadedTextures = make(map[string]*textures.Texture)
	loadedMeshes = make(map[string]*geom.Mesh)
	loadedFonts = make(map[string]*fonts.Font)
}

var (
	// A quad on the XY plane centered at (0,0) with a width and height of 2.
	QuadMesh *geom.Mesh
)

// Initialize built-in assets
func Init() {
	shaders.Init()

	QuadMesh = geom.CreateMesh(geom.Vertices{
		Pos: []mgl32.Vec3{
			{-1.0, -1.0, 0.0},
			{1.0, -1.0, 0.0},
			{-1.0, 1.0, 0.0},
			{1.0, 1.0, 0.0},
		},
		TexCoord: []mgl32.Vec2{
			{0.0, 1.0},
			{1.0, 1.0},
			{0.0, 0.0},
			{1.0, 0.0},
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

func GetTexture(assetPath string) *textures.Texture {
	texture, ok := loadedTextures[assetPath]
	if !ok {
		loadedTextures[assetPath] = textures.LoadTexture(assetPath)
		return loadedTextures[assetPath]
	}
	return texture
}

// Releases memory for all cached textures
func FreeTextures() {
	for p := range loadedTextures {
		loadedTextures[p].Free()
		delete(loadedTextures, p)
	}
}

// Releases memory for all cached meshes
func FreeMeshes() {
	for p := range loadedMeshes {
		loadedMeshes[p].Free()
		delete(loadedMeshes, p)
	}
}

// Releases memory for all cached fonts
func FreeFonts() {
	clear(loadedFonts)
}

func FreeAll() {
	FreeTextures()
	FreeMeshes()
	FreeFonts()
	FreeBuiltInAssets()
}

// Takes ownership of an already loaded mesh. Will dispose of its resources along with the other meshes.
func TakeMesh(assetPath string, mesh *geom.Mesh) {
	loadedMeshes[assetPath] = mesh
}

func GetMesh(assetPath string) (*geom.Mesh, error) {
	var err error

	mesh, ok := loadedMeshes[assetPath]
	if !ok {
		switch {
		case strings.HasSuffix(assetPath, ".obj"):
			mesh, err = geom.LoadOBJMesh(assetPath)
		default:
			err = fmt.Errorf("unsupported model file type")
		}
		if err != nil {
			return nil, err
		}
		loadedMeshes[assetPath] = mesh
	}
	return mesh, nil
}

// Retrieves a font from the game assets, loading it if it doesn't already exist.
func GetFont(assetPath string) (*fonts.Font, error) {
	var err error

	font, ok := loadedFonts[assetPath]
	if !ok {
		// Check file extension
		if !strings.HasSuffix(assetPath, ".fnt") {
			return nil, fmt.Errorf("unsupported font file type")
		}

		font, err = fonts.LoadAngelcodeFont(assetPath)
		if err != nil {
			return nil, err
		}

		loadedFonts[assetPath] = font
	}
	return font, nil
}
