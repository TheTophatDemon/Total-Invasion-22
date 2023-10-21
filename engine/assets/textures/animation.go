package textures

import "tophatdemon.com/total-invasion-ii/engine/math2"

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
