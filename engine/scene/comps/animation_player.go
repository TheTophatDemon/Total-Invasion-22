package comps

import (
	"math/rand"

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

// Changes the animation to the given one, resetting the animation counter.
func (ap *AnimationPlayer) ChangeAnimation(newAnim textures.Animation) {
	ap.animation = newAnim
	ap.currentIndex = 0
	ap.frameTimer = 0.0
}

func (ap *AnimationPlayer) CurrentAnimation() textures.Animation {
	return ap.animation
}

// Changes the animation to the given one without resetting the frame counter.
// If the animations have mismatched lengths, then the frame counter is wrapped around.
func (ap *AnimationPlayer) SwapAnimation(newAnim textures.Animation) {
	ap.animation = newAnim
	ap.currentIndex %= len(newAnim.Frames)
}

func (ap *AnimationPlayer) Update(deltaTime float32) {
	if !ap.playing || ap.animation.Frames == nil {
		return
	}
	ap.frameTimer += deltaTime
	if ap.frameTimer > ap.animation.Frames[ap.currentIndex].Duration {
		ap.currentIndex += 1
		if ap.currentIndex >= len(ap.animation.Frames) {
			if !ap.animation.Loop {
				ap.currentIndex = len(ap.animation.Frames) - 1
				ap.playing = false
			} else {
				ap.frameTimer = 0.0
				ap.currentIndex = 0
			}
		} else {
			ap.frameTimer = 0.0
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

// Returns the current frame's position as UV coordinates in the range of [0, 1)
func (ap *AnimationPlayer) FrameUV() math2.Rect {
	if ap.animation.Frames == nil || ap.currentIndex >= len(ap.animation.Frames) {
		return math2.Rect{
			X: 0.0, Y: 1.0, Width: 1.0, Height: 1.0,
		}
	}
	pixelRect := ap.animation.Frames[ap.currentIndex].Rect
	return math2.Rect{
		X:      pixelRect.X / float32(ap.animation.AtlasSize[0]),
		Y:      1.0 - (pixelRect.Y / float32(ap.animation.AtlasSize[1])),
		Width:  pixelRect.Width / float32(ap.animation.AtlasSize[0]),
		Height: pixelRect.Height / float32(ap.animation.AtlasSize[1]),
	}
}

func (ap *AnimationPlayer) MoveToRandomFrame() {
	ap.currentIndex = rand.Int() % len(ap.animation.Frames)
	ap.frameTimer = 0.0
}

func (ap *AnimationPlayer) Play() {
	ap.playing = true
}

func (ap *AnimationPlayer) Stop() {
	ap.playing = false
}

func (ap *AnimationPlayer) PlayFromStart() {
	ap.currentIndex = 0
	ap.frameTimer = 0.0
	ap.Play()
}

func (ap *AnimationPlayer) PlayNewAnim(newAnim textures.Animation) {
	ap.ChangeAnimation(newAnim)
	ap.PlayFromStart()
}

func (ap *AnimationPlayer) IsPlaying() bool {
	return ap.playing
}

func (ap *AnimationPlayer) IsAtEnd() bool {
	return ap.currentIndex == len(ap.animation.Frames)-1 && ap.frameTimer > ap.animation.Frames[ap.currentIndex].Duration
}

func (ap *AnimationPlayer) IsPlayingAnim(anim textures.Animation) bool {
	return ap.animation.BaseName() == anim.BaseName()
}

// Returns true if the animation player passed the specified trigger frame in its most recent update.
func (ap *AnimationPlayer) HitTriggerFrame(triggerFrameNumber int) bool {
	if ap.animation.TriggerFrames == nil || triggerFrameNumber >= len(ap.animation.TriggerFrames) || triggerFrameNumber < 0 {
		return false
	}
	return uint(ap.currentIndex) == ap.animation.TriggerFrames[triggerFrameNumber] && ap.frameTimer == 0.0
}

// Returns true if the animation player passed any trigger frame in its most recent update.
func (ap *AnimationPlayer) HitATriggerFrame() bool {
	for _, triggerFrame := range ap.animation.TriggerFrames {
		if uint(ap.currentIndex) == triggerFrame && ap.frameTimer == 0.0 {
			return true
		}
	}
	return false
}
