package textures

import (
	"fmt"
	_ "image/png"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Texture struct {
	target     uint32               // OpenGL Texture Target (GL_TEXTURE_2D & etc.)
	glID       uint32               // OpenGL Texture ID
	glUnit     uint32               // Texture unit (gl.TEXTURE0 for regular, gl.TEXTURE1 for atlas)
	width      uint32               // Size of entire texture
	height     uint32               // Size of the entire texture
	flags      []string             // Flags indicate the in-game properties of the texture
	animations map[string]Animation // Map of animations by name
	layers     map[string]Layer
}

type Layer struct {
	Name             string
	ViewRange        [2]int // The range of yaw angles at which this layer will be shown, in degrees.
	FlippedViewRange [2]int // The range of yaw angles at which this layer will be shown flipped horizontally, in degrees.
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

func (t *Texture) Unit() uint32 {
	return t.glUnit
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

func (tex *Texture) Rect() math2.Rect {
	return math2.Rect{
		X:      0.0,
		Y:      0.0,
		Width:  float32(tex.Width()),
		Height: float32(tex.Height()),
	}
}

func (t *Texture) Free() {
	id := t.glID
	gl.DeleteTextures(1, &id)
}

func (at *Texture) GetAnimation(name string) (anim Animation, ok bool) {
	anim, ok = at.animations[name]
	return
}

func (at *Texture) GetFirstAnimation() (Animation, error) {
	for _, anim := range at.animations {
		return anim, nil
	}
	return Animation{}, fmt.Errorf("attempted to load the first animation from a texture that has none")
}

func (t *Texture) IsAtlas() bool {
	return t.animations != nil && len(t.animations) > 0
}

func (t *Texture) Bind() {
	gl.ActiveTexture(t.glUnit)
	gl.BindTexture(t.target, t.glID)
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
