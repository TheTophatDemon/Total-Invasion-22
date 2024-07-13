package textures

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

// Loads an image from a file and flips it to be loaded into a texture.
func loadImage(assetPath string) (*image.RGBA, error) {
	if strings.ToLower(filepath.Ext(assetPath)) != ".png" {
		return nil, fmt.Errorf("cannot load %s; only .png images are supported", assetPath)
	}

	// Load image
	imgFile, err := assets.GetFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("image not found at %s", assetPath)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("error decoding image at %s", assetPath)
	}

	// Convert image to RGBA
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return nil, fmt.Errorf("error converting image at %s", assetPath)
	}
	// Flip vertically
	for x := 0; x < rgba.Bounds().Dx(); x++ {
		for y := 0; y < rgba.Bounds().Dy(); y++ {
			rgba.Set(x, y, img.At(x, rgba.Bounds().Dy()-y-1))
		}
	}

	return rgba, nil
}

func LoadTexture(assetPath string) (*Texture, error) {

	// Look for metadata file
	metaPath := strings.TrimSuffix(assetPath, ".png") + ".json"
	metadata, err := assets.LoadAndUnmarshalJSON[aseSpriteSheet](metaPath)
	if _, ok := err.(*os.PathError); err != nil && !ok {
		// The file is optional, so print errors that aren't 'file not found'.
		log.Printf("Could not parse metadata for %s; %s\n", assetPath, err)
	}

	// Load atlas image file listed in the metadata, or else use the image itself.
	var img *image.RGBA
	if metadata != nil {
		img, err = loadImage(filepath.Join(filepath.Dir(assetPath), metadata.atlasPath()))
	} else {
		img, err = loadImage(assetPath)
	}
	if err != nil {
		return ErrorTexture(), err
	}

	texture := &Texture{
		width:  uint32(img.Bounds().Dx()),
		height: uint32(img.Bounds().Dy()),
	}

	if metadata != nil {
		texture.flags = metadata.loadFlags()

		if texture.animations, err = metadata.loadAnimations(); err != nil {
			log.Printf("Could not load animations for %s; %s\n", assetPath, err)
		}
		if texture.layers, err = metadata.loadLayers(); err != nil {
			log.Printf("Could not load layers for %s; %s\n", assetPath, err)
		}
		if metadata.Meta.Slices != nil {
			texture.slices = make(map[string]Slice)
			for _, aseSlice := range metadata.Meta.Slices {
				if aseSlice.Name != META_SLICE_NAME {
					texture.slices[aseSlice.Name] = Slice{
						Data: aseSlice.Data,
						Bounds: math2.Rect{
							X:      float32(aseSlice.Keys[0].Bounds.X),
							Y:      float32(aseSlice.Keys[0].Bounds.Y),
							Width:  float32(aseSlice.Keys[0].Bounds.W),
							Height: float32(aseSlice.Keys[0].Bounds.H),
						},
					}
				}
			}
		}
	}

	// Set texture data as whole image
	texture.target = gl.TEXTURE_2D
	texture.glUnit = gl.TEXTURE0
	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &texture.glID)
	gl.BindTexture(texture.target, texture.glID)
	gl.TexImage2D(texture.target, 0, gl.RGBA, int32(texture.width), int32(texture.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

	// Apply filtering and mipmapping
	gl.TexParameteri(texture.Target(), gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_NEAREST)
	if texture.HasFlag(FLAG_CLAMP_BORDER) {
		gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_S, gl.CLAMP_TO_BORDER)
		gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_T, gl.CLAMP_TO_BORDER)
	} else {
		gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_S, gl.REPEAT)
		gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_T, gl.REPEAT)
	}
	gl.GenerateMipmap(texture.Target())

	log.Printf("Texture loaded at %v.\n", assetPath)

	return texture, nil
}
