package textures

import (
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type Animation struct {
	Frames    []Frame
	AtlasSize [2]uint
	Name      string
	Loop      bool
}

type Frame struct {
	Rect     math2.Rect // Source rectangle on the texture atlas
	Duration float32    // Duration of the frame in seconds
}

func (anim *Animation) BaseName() string {
	semiIndex := strings.LastIndexByte(anim.Name, byte(';'))
	if semiIndex > -1 {
		return anim.Name[:semiIndex]
	}
	return anim.Name
}

func (anim *Animation) LayerName() string {
	semiIndex := strings.LastIndexByte(anim.Name, byte(';'))
	if semiIndex > -1 && semiIndex < len(anim.Name)-1 {
		return anim.Name[semiIndex+1:]
	}
	return ""
}

// Returns the number of seconds from the start to the end.
func (anim *Animation) Duration() float32 {
	var sum float32
	for _, frame := range anim.Frames {
		sum += frame.Duration
	}
	return sum
}
