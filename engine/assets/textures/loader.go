package textures

import (
	"fmt"
	"image"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

// Loads an image from a file and flips it to be loaded into a texture.
func loadImage(assetPath string) (*image.RGBA, error) {
	if strings.ToLower(path.Ext(assetPath)) != ".png" {
		return nil, fmt.Errorf("cannot load %s; only .png images are supported", assetPath)
	}

	//Load image
	imgFile, err := assets.GetFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("error loading image at %s", assetPath)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("error decoding image at %s", assetPath)
	}

	//Convert image to RGBA
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return nil, fmt.Errorf("error converting image at %s", assetPath)
	}
	//Flip vertically
	for x := 0; x < rgba.Bounds().Dx(); x++ {
		for y := 0; y < rgba.Bounds().Dy(); y++ {
			rgba.Set(x, y, img.At(x, rgba.Bounds().Dy()-y-1))
		}
	}

	return rgba, nil
}

// Creates a new image representing the rectangle subsection of the source image.
func subImageCopy(src *image.RGBA, x, y, w, h int) *image.RGBA {
	dest := image.NewRGBA(image.Rect(0, 0, w, h))

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			dest.Set(i, j, src.At(i+x, j+y))
		}
	}

	return dest
}

func LoadTexture(assetPath string) *Texture {

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
		img, err = loadImage(path.Join(path.Dir(assetPath), metadata.atlasPath()))
	} else {
		img, err = loadImage(assetPath)
	}
	if err != nil {
		log.Println(err)
		return ErrorTexture()
	}

	texture := &Texture{
		width:  uint32(img.Bounds().Dx()),
		height: uint32(img.Bounds().Dy()),
	}

	if metadata != nil {
		if texture.flags, err = metadata.loadFlags(); err != nil {
			log.Printf("Could not parse texture flags for %s; %s\n", assetPath, err)
		}
		if texture.animations, err = metadata.loadAnimations(); err != nil {
			log.Printf("Could not load animations for %s; %s\n", assetPath, err)
		}
		if texture.layers, err = metadata.loadLayers(); err != nil {
			log.Printf("Could not load layers for %s; %s\n", assetPath, err)
		}
	}

	//Set texture data as whole image
	texture.target = gl.TEXTURE_2D
	texture.glUnit = gl.TEXTURE0
	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &texture.glID)
	gl.BindTexture(texture.target, texture.glID)
	gl.TexImage2D(texture.target, 0, gl.RGBA, int32(texture.width), int32(texture.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))

	//Apply filtering and mipmapping
	gl.TexParameteri(texture.Target(), gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_NEAREST)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.GenerateMipmap(texture.Target())

	log.Printf("Texture loaded at %v.\n", assetPath)

	return texture
}
