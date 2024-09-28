package textures

import (
	_ "image/png"
	"strings"

	"github.com/go-gl/gl/v3.3-core/gl"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

const (
	FLAG_CLAMP_BORDER = "clampBorder"
)

type Texture struct {
	target     uint32               // OpenGL Texture Target (GL_TEXTURE_2D & etc.)
	glID       uint32               // OpenGL Texture ID
	glUnit     uint32               // Texture unit (gl.TEXTURE0 for regular, gl.TEXTURE1 for atlas)
	width      uint32               // Size of entire texture
	height     uint32               // Size of the entire texture
	flags      []string             // Flags indicate the in-game properties of the texture
	slices     map[string]Slice     // Holds the slices defined in Aseprite (excluding the meta slice). Indexed by name.
	animations map[string]Animation // Map of animations by name
	layers     map[string]Layer
}

type Layer struct {
	Name             string
	ViewRange        [2]int // The range of yaw angles at which this layer will be shown, in degrees.
	FlippedViewRange [2]int // The range of yaw angles at which this layer will be shown flipped horizontally, in degrees.
}

type Slice struct {
	Data   string
	Bounds math2.Rect
}

func (tex *Texture) Width() int {
	return int(tex.width)
}

func (tex *Texture) Height() int {
	return int(tex.height)
}

func (tex *Texture) ID() uint32 {
	return tex.glID
}

func (tex *Texture) Target() uint32 {
	return tex.target
}

func (tex *Texture) Unit() uint32 {
	return tex.glUnit
}

// Returns true if the texture has a flag matching the argument (ignoring case).
func (tex *Texture) HasFlag(testFlag string) bool {
	for f := range tex.flags {
		if strings.EqualFold(tex.flags[f], testFlag) {
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

func (tex *Texture) Free() {
	id := tex.glID
	gl.DeleteTextures(1, &id)
}

func (tex *Texture) GetDefaultAnimation() Animation {
	for _, anim := range tex.animations {
		if anim.Default || len(tex.animations) == 1 {
			return anim
		}
	}
	return Animation{}
}

func (tex *Texture) AnimationCount() int {
	return len(tex.animations)
}

func (tex *Texture) GetAnimation(name string) (anim Animation, ok bool) {
	anim, ok = tex.animations[name]
	return
}

func (tex *Texture) GetAnimationNames() []string {
	result := make([]string, 0, len(tex.animations))
	for name := range tex.animations {
		result = append(result, name)
	}
	return result
}

func (tex *Texture) LayerCount() int {
	return len(tex.layers)
}

// Returns the slice with the given name from the texture.
// If it is not found, a zero-value slice is returned.
func (tex *Texture) FindSlice(name string) Slice {
	return tex.slices[name]
}

// Searches for a layer in this texture with the given degrees angle in its range.
//
// The first boolean returned indicates whether the angle is within the flipped view range.
//
// The second boolean indicates whether the angle is within either view range.
func (tex *Texture) FindLayerWithinAngle(angle int) (Layer, bool, bool) {
	angle %= 360
	if angle < 0 {
		angle += 360
	}
	for l := range tex.layers {
		for a := angle; a >= angle-360; a -= 360 {
			if a >= tex.layers[l].ViewRange[0] && a < tex.layers[l].ViewRange[1] {
				return tex.layers[l], false, true
			}
			if a >= tex.layers[l].FlippedViewRange[0] && a < tex.layers[l].FlippedViewRange[1] {
				return tex.layers[l], true, true
			}
		}
	}
	return Layer{}, false, false
}

func (tex *Texture) IsAtlas() bool {
	return tex.animations != nil && len(tex.animations) > 0
}

func (tex *Texture) Bind() {
	gl.ActiveTexture(tex.glUnit)
	gl.BindTexture(tex.target, tex.glID)
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
