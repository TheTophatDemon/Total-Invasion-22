package assets

import (
	"fmt"
	"image"
	_ "image/png"
	"log"
	"os"
	"path"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type Texture struct {
	target      uint32           // OpenGL Texture Target (GL_TEXTURE_2D & etc.)
	glID        uint32           // OpenGL Texture ID
	width       uint32           // Size of entire texture
	height      uint32           // Size of the entire texture
	flags       []string         // Flags indicate the in-game properties of the texture
	frameWidth  uint32           // Width of an animation frame
	frameHeight uint32           // Height of an animation frame
	animations  []FrameAnimation // List of animations (may be empty)
}

// Represents the contents of a texture configuration .json file.
type TextureMetadata struct {
	Atlas      string    // Path to the atlas texture relative to the program executable.
	FrameSize  [2]uint32 // The size of a frame in the atlas.
	Animations []FrameAnimation
	Flags      []string // Defines in-game properties of the texture.
}

func (t *Texture) Width() int {
	return int(t.width)
}

func (t *Texture) Height() int {
	return int(t.height)
}

func (t *Texture) ID() uint32 {
	return t.glID
}

func (t *Texture) Target() uint32 {
	return t.target
}

// Returns true if the texture has a flag matching the argument (ignoring case).
func (t *Texture) HasFlag(testFlag string) bool {
	for f := range t.flags {
		if strings.EqualFold(t.flags[f], testFlag) {
			return true
		}
	}
	return false
}

func (t *Texture) Free() {
	id := t.glID
	gl.DeleteTextures(1, &id)
}

func (at *Texture) GetAnimation(index int) FrameAnimation {
	return at.animations[index]
}

func (at *Texture) AnimationCount() int {
	return len(at.animations)
}

func (at *Texture) FrameWidth() int {
	return int(at.frameWidth)
}

func (at *Texture) FrameHeight() int {
	return int(at.frameHeight)
}

func (t *Texture) IsAtlas() bool {
	return t.animations != nil && len(t.animations) > 0
}

const ERROR_TEXTURE_SIZE = 64

var errorTexture *Texture

// Returns and/or generates the error texture, a magenta-and-black checkered image.
func ErrorTexture() *Texture {
	if errorTexture == nil {
		errorTexture = new(Texture)
		gl.GenTextures(1, &errorTexture.glID)
		gl.ActiveTexture(gl.TEXTURE0)
		gl.BindTexture(gl.TEXTURE_2D, errorTexture.glID)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

		data := make([]uint8, 0, ERROR_TEXTURE_SIZE*ERROR_TEXTURE_SIZE)

		for x := 0; x < ERROR_TEXTURE_SIZE; x++ {
			for y := 0; y < ERROR_TEXTURE_SIZE; y++ {
				isBlack := false
				if ((x/16)%2 == 0) && ((y/16)%2 == 0) {
					isBlack = true
				} else if ((x/16)%2 == 1) && ((y/16)%2 == 1) {
					isBlack = true
				}

				if isBlack {
					data = append(data, 0, 0, 0, 255)
				} else {
					data = append(data, 255, 0, 255, 255)
				}
			}
		}

		gl.TexImage2D(
			gl.TEXTURE_2D, 0, gl.RGBA, ERROR_TEXTURE_SIZE, ERROR_TEXTURE_SIZE, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
	}
	return errorTexture
}

// Loads an image from a file and flips it to be loaded into a texture.
func loadImage(assetPath string) (*image.RGBA, error) {
	if strings.ToLower(path.Ext(assetPath)) != ".png" {
		return nil, fmt.Errorf("cannot load %s; only .png images are supported", assetPath)
	}

	//Load image
	imgFile, err := getFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("error loading texture at %s", assetPath)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("error decoding texture at %s", assetPath)
	}

	//Convert image to RGBA
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return nil, fmt.Errorf("error converting texture at %s", assetPath)
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

func loadTexture(assetPath string) *Texture {

	//Look for metadata file
	metaPath := strings.TrimSuffix(assetPath, ".png") + ".json"
	metadata, err := LoadAndUnmarshalJSON[TextureMetadata](metaPath)
	if _, ok := err.(*os.PathError); err != nil && !ok {
		//The file is optional, so print errors that aren't 'file not found'.
		log.Printf("Could not parse metadata for %s; %s\n", assetPath, err)
	}

	//Load atlas image file listed in the metadata, or else use the image itself.
	var img *image.RGBA
	if metadata != nil && len(metadata.Atlas) > 0 {
		img, err = loadImage(metadata.Atlas)
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

	//Generate OpenGL Texture
	if metadata != nil && len(metadata.Atlas) > 0 && metadata.FrameSize[0] > 0 && metadata.FrameSize[1] > 0 {
		// If it's an atlas texture...
		texture.target = gl.TEXTURE_2D_ARRAY
		texture.frameWidth = metadata.FrameSize[0]
		texture.frameHeight = metadata.FrameSize[1]
		texture.animations = make([]FrameAnimation, len(metadata.Animations))
		copy(texture.animations, metadata.Animations)

		cols := int(texture.width / texture.frameWidth)
		rows := int(texture.height / texture.frameHeight)
		nFrames := rows * cols

		gl.ActiveTexture(gl.TEXTURE1)
		gl.GenTextures(1, &texture.glID)
		gl.BindTexture(texture.target, texture.glID)

		gl.TexImage3D(texture.target, 0, gl.RGBA, int32(texture.frameWidth), int32(texture.frameHeight), int32(nFrames), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

		// Generate subimages for each frame
		for x := 0; x < cols; x++ {
			for y := 0; y < rows; y++ {
				ofsx := x * texture.FrameWidth()
				ofsy := y * texture.FrameHeight()
				frameRGBA := subImageCopy(img, ofsx, ofsy, texture.FrameWidth(), texture.FrameHeight())

				gl.TexSubImage3D(texture.target, 0, 0, 0,
					int32(x+y*cols),
					int32(texture.frameWidth),
					int32(texture.frameHeight),
					1,
					gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(frameRGBA.Pix))
			}
		}
	} else {
		//Set texture data as whole image
		texture.target = gl.TEXTURE_2D
		texture.frameWidth = texture.width
		texture.frameHeight = texture.height
		texture.animations = nil
		gl.ActiveTexture(gl.TEXTURE0)
		gl.GenTextures(1, &texture.glID)
		gl.BindTexture(texture.target, texture.glID)
		gl.TexImage2D(texture.target, 0, gl.RGBA, int32(texture.width), int32(texture.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	}

	//Apply filtering and mipmapping
	gl.TexParameteri(texture.Target(), gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_NEAREST)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.GenerateMipmap(texture.Target())

	log.Println("Texture loaded at ", assetPath, ".")

	return texture
}
