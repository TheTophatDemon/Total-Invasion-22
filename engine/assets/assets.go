package assets

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
)

// This map caches the textures loaded from the filesystem by their paths.
var textures map[string]*Texture

// This caches the meshes loaded from the filesystem by their paths.
var meshes map[string]*Mesh

func init() {
	textures = make(map[string]*Texture)
	meshes = make(map[string]*Mesh)
}

var (
	MapShader *Shader

	//go:embed map.vs.glsl
	mapVertShaderSrc string
	//go:embed map.fs.glsl
	mapFragShaderSrc string

	SpriteShader *Shader

	//go:embed sprite.vs.glsl
	spriteVertShaderSrc string
	//go:embed sprite.fs.glsl
	spriteFragShaderSrc string

	SpriteMesh *Mesh
)

// Initialize built-in assets
func Init() {
	var err error

	MapShader, err = CreateShader(mapVertShaderSrc, mapFragShaderSrc)
	if err != nil {
		log.Fatalln("Couldn't compile map shader: ", err)
	}

	SpriteShader, err = CreateShader(spriteVertShaderSrc, spriteFragShaderSrc)
	if err != nil {
		log.Fatalln("Couldn't compile sprite shader: ", err)
	}

	SpriteMesh = CreateMesh(Vertices{
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
	MapShader.Free()
	SpriteShader.Free()
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

func FreeAll() {
	FreeTextures()
	FreeMeshes()
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
			return nil, nil
		}
		meshes[assetPath] = mesh
	}
	return mesh, nil
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
