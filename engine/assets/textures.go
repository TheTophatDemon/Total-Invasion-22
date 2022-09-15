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
	"tophatdemon.com/total-invasion-ii/engine"
)

type TextureFlag uint32

const (
	TEXFLAG_NONE     TextureFlag = 0 << 0
	TEXFLAG_UNSHADED TextureFlag = 1 << 0
	TEXFLAG_LIQUID   TextureFlag = 1 << 1
)

var textureFlagFromString = map[string]TextureFlag {
	"unshaded": TEXFLAG_UNSHADED,
	"liquid"  : TEXFLAG_LIQUID,
}

type Texture interface {
	Width()              int
	Height()             int
	ID()                 uint32
	Target()             uint32
	HasFlag(TextureFlag) bool
}

type BaseTexture struct {
	target      uint32
	glID        uint32
	width       uint32 //Size of entire texture
	height      uint32
	flags       TextureFlag
}

type AtlasTexture struct {
	BaseTexture
	frameWidth  uint32
	frameHeight uint32
	animations  []FrameAnimation
}

type TextureMetadata struct {
	Atlas      string
	FrameSize  [2]uint32
	Animations []FrameAnimation
	Flags      []string
}

func (t *BaseTexture) Width() int {
	return int(t.width)
}

func (t *BaseTexture) Height() int {
	return int(t.height)
}

func (t *BaseTexture) ID() uint32 {
	return t.glID
}

func (t *BaseTexture) Target() uint32 {
	return t.target
}

func (t *BaseTexture) HasFlag(testFlag TextureFlag) bool {
	return (t.flags & testFlag) == testFlag
}

func (at *AtlasTexture) GetAnimation(index int) FrameAnimation {
	return at.animations[index]
}

func (at *AtlasTexture) AnimationCount() int {
	return len(at.animations)
}

func (at *AtlasTexture) FrameWidth() int {
	return int(at.frameWidth)
}

func (at *AtlasTexture) FrameHeight() int {
	return int(at.frameHeight)
}

const ERROR_TEXTURE_SIZE = 64

var errorTexture *BaseTexture

//Returns and/or generates the error texture, a magenta-and-black checkered image.
func ErrorTexture() Texture {
	if errorTexture == nil {
		errorTexture = new(BaseTexture)
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

//Loads an image from a file and flips it to be loaded into a texture.
func loadImage(assetPath string) (*image.RGBA, error) {
	if strings.ToLower(path.Ext(assetPath)) != ".png" {
		return nil, fmt.Errorf("Cannot load %s; only .png images are supported!", assetPath)
	}

	//Load image
	imgFile, err := getFile(assetPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading texture at %s.", assetPath)
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("Error decoding texture at %s.", assetPath)
	}
	
	//Convert image to RGBA
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return nil, fmt.Errorf("Error converting texture at %s.", assetPath)
	}
	//Flip vertically
	for x := 0; x < rgba.Bounds().Dx(); x++ {
		for y := 0; y < rgba.Bounds().Dy(); y++ {
			rgba.Set(x, y, img.At(x, rgba.Bounds().Dy() - y - 1))
		}
	}

	return rgba, nil
}

//Creates a new image representing the rectangle subsection of the source image.
func subImageCopy(src *image.RGBA, x, y, w, h int) *image.RGBA {
	dest := image.NewRGBA(image.Rect(0, 0, w, h))

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			dest.Set(i, j, src.At(i + x, j + y))
		}
	}

	return dest
}

func loadTexture(assetPath string) Texture {

	//Look for metadata file
	metaPath := strings.TrimSuffix(assetPath, ".png") + ".json"
	metadata, err := LoadAndUnmarshalJSON[TextureMetadata](metaPath)
	if _, ok := err.(*os.PathError); err != nil && !ok {
		//The file is optional, so print errors that aren't 'file not found'.
		log.Printf("Could not parse metadata for %s; %s\n", assetPath, err)
	}
	
	var texture Texture

	//Generate OpenGL Texture
	engine.CheckOpenGLError()
	engine.CheckOpenGLError()
	if metadata != nil && metadata.FrameSize[0] > 0 && metadata.FrameSize[1] > 0 {
		//Providing a frame size turns it into an atlas texture

		//Load atlas file listed in the metadata, or else use the image itself.
		var atlasImg *image.RGBA
		if len(metadata.Atlas) > 0 {
			atlasImg, err = loadImage(metadata.Atlas)
		} else {
			atlasImg, err = loadImage(assetPath)
		}
		if err != nil {
			log.Println(err)
			return ErrorTexture()
		}

		atlasTexture := &AtlasTexture{
			BaseTexture: BaseTexture{
				target: gl.TEXTURE_2D_ARRAY,
				width: uint32(atlasImg.Bounds().Dx()),
				height: uint32(atlasImg.Bounds().Dy()),
			},
			frameWidth: metadata.FrameSize[0],
			frameHeight: metadata.FrameSize[1],
		}

		rows := int(atlasTexture.width / atlasTexture.frameWidth)
		cols := int(atlasTexture.height / atlasTexture.frameHeight)
		nFrames := rows * cols
		
		engine.CheckOpenGLError()

		gl.ActiveTexture(gl.TEXTURE1)
		gl.GenTextures(1, &atlasTexture.glID)
		gl.BindTexture(atlasTexture.target, atlasTexture.glID)
		
		engine.CheckOpenGLError()

		gl.TexImage3D(atlasTexture.target, 0, gl.RGBA, int32(atlasTexture.frameWidth), int32(atlasTexture.frameHeight), int32(nFrames), 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

		engine.CheckOpenGLError()
		// Generate subimages for each frame
		for x := 0; x < cols; x++ {
			for y := 0; y < rows; y++ {
				ofsx := x * atlasTexture.FrameWidth()
				ofsy := y * atlasTexture.FrameHeight()
				frameRGBA := subImageCopy(atlasImg, ofsx, ofsy, atlasTexture.FrameWidth(), atlasTexture.FrameHeight())

				gl.TexSubImage3D(atlasTexture.target, 0, 0, 0, 
					int32(x + y * cols), 
					int32(atlasTexture.frameWidth), 
					int32(atlasTexture.frameHeight), 
					1, 
					gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(frameRGBA.Pix))
			}
		}
		engine.CheckOpenGLError()

		texture = atlasTexture
	} else {
		//Set texture data as whole image

		rgba, err := loadImage(assetPath)
		if err != nil {
			log.Println(err)
			return ErrorTexture()
		}

		baseTexture := &BaseTexture{
			target: gl.TEXTURE_2D,
			width: uint32(rgba.Bounds().Dx()), 
			height: uint32(rgba.Bounds().Dy()),
		}

		engine.CheckOpenGLError()
		gl.ActiveTexture(gl.TEXTURE0)
		gl.GenTextures(1, &baseTexture.glID)
		gl.BindTexture(baseTexture.target, baseTexture.glID)
		gl.TexImage2D(baseTexture.target, 0, gl.RGBA, int32(baseTexture.width), int32(baseTexture.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
		engine.CheckOpenGLError()

		texture = baseTexture
	}

	//Apply filtering and mipmapping
	engine.CheckOpenGLError()
	gl.TexParameteri(texture.Target(), gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_NEAREST)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(texture.Target(), gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.GenerateMipmap(texture.Target())

	log.Println("Texture loaded at ", assetPath, ".")

	engine.CheckOpenGLError()

	return texture
}