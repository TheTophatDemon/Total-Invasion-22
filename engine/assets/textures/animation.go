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

func (a *Animation) BaseName() string {
	semiIndex := strings.LastIndexByte(a.Name, byte(';'))
	if semiIndex > -1 {
		return a.Name[:semiIndex]
	}
	return a.Name
}

func (a *Animation) LayerName() string {
	semiIndex := strings.LastIndexByte(a.Name, byte(';'))
	if semiIndex > -1 && semiIndex < len(a.Name)-1 {
		return a.Name[semiIndex+1:]
	}
	return ""
}
