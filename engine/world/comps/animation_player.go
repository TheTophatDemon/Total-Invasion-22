package comps

import (
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
	"tophatdemon.com/total-invasion-ii/engine/math2"
)

type AnimationPlayer struct {
	animation    textures.Animation // Animation currently being played
	currentIndex int                // The current index into the animation data's frames array
	playing      bool
	frameTimer   float32
}

func NewAnimationPlayer(anim textures.Animation, autoPlay bool) AnimationPlayer {
	return AnimationPlayer{
		animation:    anim,
		currentIndex: 0,
		playing:      autoPlay,
		frameTimer:   0.0,
	}
}

func (ap *AnimationPlayer) ChangeAnimation(newAnim textures.Animation) {
	ap.animation = newAnim
	ap.currentIndex = 0
	ap.frameTimer = 0.0
}

func (ap *AnimationPlayer) Update(deltaTime float32) {
	if !ap.playing || ap.animation.Frames == nil {
		return
	}
	ap.frameTimer += deltaTime
	if ap.frameTimer > ap.animation.Frames[ap.currentIndex].Duration {
		ap.frameTimer = 0.0
		ap.currentIndex += 1
		if ap.currentIndex >= len(ap.animation.Frames) {
			if !ap.animation.Loop {
				ap.currentIndex = len(ap.animation.Frames) - 1
				ap.playing = false
			} else {
				ap.currentIndex = 0
			}
		}
	}
}

func (ap *AnimationPlayer) Frame() textures.Frame {
	if ap.animation.Frames == nil || ap.currentIndex >= len(ap.animation.Frames) {
		return textures.Frame{
			Rect: math2.Rect{
				X: 0.0, Y: 0.0,
				Width:  float32(ap.animation.AtlasSize[0]),
				Height: float32(ap.animation.AtlasSize[1]),
			},
			Duration: 0.0,
		}
	}
	return ap.animation.Frames[ap.currentIndex]
}

func (ap *AnimationPlayer) Play() {
	ap.playing = true
	ap.frameTimer = 0.0
}
