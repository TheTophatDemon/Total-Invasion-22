package cache

import (
	"fmt"
	"log"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/assets/audio"
	"tophatdemon.com/total-invasion-ii/engine/assets/fonts"
	"tophatdemon.com/total-invasion-ii/engine/assets/geom"
	"tophatdemon.com/total-invasion-ii/engine/assets/locales"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/failure"
	"tophatdemon.com/total-invasion-ii/engine/iter"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
)

type cache[T any] struct {
	storage        map[string]T
	fileExtensions []string
	loadFunc       func(string) (T, error)
	freeFunc       func(T)
	resourceName   string
}

// This map caches the loadedTextures loaded from the filesystem by their paths.
var loadedTextures cache[*textures.Texture]

// This caches the meshes loaded from the filesystem by their paths.
var loadedMeshes cache[*geom.Mesh]

// Cache of bitmap fonts, indexed by .fnt file path
var loadedFonts cache[*fonts.Font]

// Cache of sound effects, indexed by .wav file path
var loadedSfx cache[tdaudio.SoundId]

// Cache of translations, indexed by .json file path.
var loadedTranslations cache[*locales.Translation]

func init() {
	loadedTextures = cache[*textures.Texture]{
		storage:        make(map[string]*textures.Texture),
		fileExtensions: []string{".png"},
		loadFunc:       textures.LoadTexture,
		freeFunc:       (*textures.Texture).Free,
		resourceName:   "texture",
	}
	loadedMeshes = cache[*geom.Mesh]{
		storage:        make(map[string]*geom.Mesh),
		fileExtensions: []string{".obj"},
		loadFunc:       geom.LoadOBJMesh,
		freeFunc:       (*geom.Mesh).Free,
		resourceName:   "mesh",
	}
	loadedFonts = cache[*fonts.Font]{
		storage:        make(map[string]*fonts.Font),
		fileExtensions: []string{".fnt"},
		loadFunc:       fonts.LoadAngelcodeFont,
		freeFunc:       nil,
		resourceName:   "font",
	}
	loadedSfx = cache[tdaudio.SoundId]{
		storage:        make(map[string]tdaudio.SoundId),
		fileExtensions: []string{".wav"},
		loadFunc:       audio.LoadSfx,
		freeFunc:       nil,
		resourceName:   "sfx",
	}
	loadedTranslations = cache[*locales.Translation]{
		storage:        make(map[string]*locales.Translation),
		fileExtensions: []string{".json"},
		loadFunc:       locales.LoadTranslation,
		freeFunc:       nil,
		resourceName:   "translation",
	}
}

func (c *cache[T]) get(assetPath string) (T, error) {
	var err error
	var empty T
	assetPath = strings.ReplaceAll(assetPath, "\\", "/")
	resource, ok := c.storage[assetPath]
	if !ok {
		for _, extension := range c.fileExtensions {
			if strings.HasSuffix(assetPath, extension) {
				resource, err = c.loadFunc(assetPath)
				if err != nil {
					return empty, err
				}
				c.storage[assetPath] = resource
				return resource, nil
			}
		}
		return empty, fmt.Errorf("unsupported file type for %v", c.resourceName)
	}
	return resource, nil
}

func (c *cache[T]) freeAll() {
	if c.freeFunc != nil {
		for i := range c.storage {
			c.freeFunc(c.storage[i])
		}
	}
	clear(c.storage)
}

func (c *cache[T]) take(assetPath string, resource T) {
	assetPath = strings.ReplaceAll(assetPath, "\\", "/")
	c.storage[assetPath] = resource
}

func (c *cache[T]) iterate() iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		for path, item := range c.storage {
			if !yield(path, item) {
				return
			}
		}
	}
}

func GetTexture(assetPath string) *textures.Texture {
	texture, err := loadedTextures.get(assetPath)
	if err != nil {
		log.Println(err)
		return textures.ErrorTexture()
	}
	return texture
}

// Takes ownership of an already loaded mesh. Will dispose of its resources along with the other meshes.
func TakeMesh(assetPath string, mesh *geom.Mesh) {
	loadedMeshes.take(assetPath, mesh)
}

func GetMesh(assetPath string) (*geom.Mesh, error) {
	mesh, err := loadedMeshes.get(assetPath)
	if err != nil {
		log.Println(err)
	}
	return mesh, err
}

// Retrieves a font from the game assets, loading it if it doesn't already exist.
func GetFont(assetPath string) (*fonts.Font, error) {
	fnt, err := loadedFonts.get(assetPath)
	if err != nil {
		log.Println(err)
	}
	return fnt, err
}

// Retrieves a sound effect from the game assets.
func GetSfx(assetPath string) tdaudio.SoundId {
	sfx, err := loadedSfx.get(assetPath)
	if err != nil {
		log.Println(err)
	}
	return sfx
}

// Retrieves a translation from the game assets, loading it if iit doesn't already exist.
func GetTranslation(assetPath string) (*locales.Translation, error) {
	trans, err := loadedTranslations.get(assetPath)
	if err != nil {
		failure.LogErrWithLocation("could not load translation at %v: %v", assetPath, err)
	}
	return trans, err
}

// This frees memory for all resources currently loaded and clears the cache.
// This is done to refresh the loaded assets between levels.
func Reset() {
	loadedTextures.freeAll()
	clear(loadedTextures.storage)
	loadedMeshes.freeAll()
	clear(loadedMeshes.storage)
	loadedFonts.freeAll()
	clear(loadedFonts.storage)
}

// This frees memory for all resources currently loaded.
// This is meant to be done permanently at the end of the program.
func FreeAll() {
	loadedTextures.freeAll()
	loadedMeshes.freeAll()
	loadedFonts.freeAll()
	FreeBuiltInAssets()
}
