package assets

import (
	"image"
	_ "image/png"
	"log"

	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine"
)

type Texture struct {
	glID   uint32
	width  uint32
	height uint32
}

func (t *Texture) GetWidth() int {
	return int(t.width)
}

func (t *Texture) GetHeight() int {
	return int(t.height)
}

func (t *Texture) GetSize() (int, int) {
	return int(t.width), int(t.height)
}

func (t *Texture) GetID() uint32 {
	return t.glID
}

const ERROR_TEXTURE_SIZE = 64

var errorTexture *Texture

//Returns and/or generates the error texture, a magenta-and-black checkered image.
func GetErrorTexture() *Texture {
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

func loadTexture(assetPath string) *Texture {
	imgFile, err := getFile(assetPath)
	if err != nil {
		log.Println("Error loading texture at ", assetPath, ".")
		return GetErrorTexture()
	}
	defer imgFile.Close()

	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Println("Error decoding texture at ", assetPath, ".")
		return GetErrorTexture()
	}
	
	//Convert image to RGBA and flip vertically
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		log.Println("Error converting texture at ", assetPath, ".")
		return GetErrorTexture()
	}
	for x := 0; x < rgba.Bounds().Dx(); x++ {
		for y := 0; y < rgba.Bounds().Dy(); y++ {
			rgba.Set(x, y, img.At(x, rgba.Bounds().Dy() - y - 1))
		}
	}
	
	texture := Texture{}
	texture.width = uint32(rgba.Bounds().Dx())
	texture.height = uint32(rgba.Bounds().Dy())
	
	//Generate OpenGL Texture
	gl.GenTextures(1, &texture.glID)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture.glID)
	gl.TexImage2D(
		gl.TEXTURE_2D, 0, gl.RGBA, int32(texture.width), int32(texture.height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.GenerateMipmap(gl.TEXTURE_2D)

	log.Println("Texture loaded at ", assetPath, ".")

	engine.CheckOpenGLError()

	return &texture
}