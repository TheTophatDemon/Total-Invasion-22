package ecomps

import (
	"tophatdemon.com/total-invasion-ii/engine/assets"
	"tophatdemon.com/total-invasion-ii/engine/scene"
)

type AnimationPlayer struct {
	animation    assets.FrameAnimation //Animation currently being played
	currentFrame int                   //The current frame number from the animation data
	currentIndex int                   //The current index into the animation data's frames array
	playing      bool
	frameTimer   float32
}

func AddAnimationPlayer(ent scene.Entity, anim assets.FrameAnimation, autoPlay bool) {
	AnimationPlayerComps.Assign(ent, AnimationPlayer{
		animation:    anim,
		currentFrame: 0,
		currentIndex: 0,
		playing:      autoPlay,
		frameTimer:   0.0,
	})
}

func (ap *AnimationPlayer) ChangeAnimation(newAnim assets.FrameAnimation) {
	ap.animation = newAnim
	ap.currentFrame = newAnim.Frames[0]
	ap.currentIndex = 0
	ap.frameTimer = 0.0
}

func (ap *AnimationPlayer) Update(deltaTime float32) {
	if !ap.playing {
		return
	}
	ap.frameTimer += deltaTime
	if ap.frameTimer > ap.animation.Speed {
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
		ap.currentFrame = ap.animation.Frames[ap.currentIndex]
	}
}

func (ap *AnimationPlayer) Frame() int {
	return ap.currentFrame
}

func (ap *AnimationPlayer) Play() {
	ap.playing = true
	ap.frameTimer = 0.0
	ap.currentFrame = 0
}
