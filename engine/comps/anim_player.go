package comps

// import "tophatdemon.com/total-invasion-ii/engine/assets"

// type AnimationPlayer struct {
// 	animation    assets.FrameAnimation
// 	currentFrame int //The current frame number from the animation data
// 	currentIndex int //The current index into the animation data's frames array
// 	playing      bool 
// 	frameTimer   float32
// }

// //Resets the state of the animation player to assume the animation set of the given texture atlas.
// func (ap *AnimationPlayer) FromFrames(atlas *assets.AtlasTexture, animIndex int) {
// 	//Copy animations
// 	ap.animation = atlas.GetAnimation(animIndex)

// 	ap.currentAnim = 0
// 	ap.currentFrame = 0
// 	ap.frameTimer = 0.0	
// }

// func (ap *AnimationPlayer) Update(deltaTime float32) {
// 	if !ap.playing { return }
// 	anim := ap.animations[ap.currentAnim]
// 	ap.frameTimer += deltaTime
// 	if ap.frameTimer > anim.Speed {
// 		ap.frameTimer = 0.0
// 		ap.currentIndex += 1
// 		if ap.currentIndex >= len(anim.Frames) {
// 			if anim.Loop {
// 				ap.currentIndex = len(anim.Frames)
// 			} else {
// 				ap.currentIndex = 0
// 			}
// 		}
// 		ap.currentFrame = anim.Frames[ap.currentIndex]
// 	}
// }

// func (ap *AnimationPlayer) Frame() int {
// 	return ap.currentFrame
// }

// func (ap *AnimationPlayer) Play() {
// 	ap.playing = true
// 	ap.frameTimer = 0.0
// 	ap.currentFrame = 0
// }