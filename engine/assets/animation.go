package assets

import "tophatdemon.com/total-invasion-ii/engine/math2"

type AnimationSet struct {
	Anims  []Animation
	Layers []Layer
}

type Animation struct {
	Frames []Frame
	Name   string
	Loop   bool
}

type Frame struct {
	Rect     math2.Rect
	Duration uint32
}

type Layer struct {
	Name     string
	Metadata string // JSON metadata
}
