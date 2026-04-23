package hud

import (
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/containers/maybe"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/engine/tween"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const LevelIntroTime = 3.0 // Time after which the level intro ends.

type LevelIntro struct {
	timer                           float32
	voice                           tdaudio.VoiceId
	starBackground, blackBackground ui.Element
	bannerTop, bannerBottom         ui.Element
	sickle, eyes                    ui.Element
	sickleXSeq, sickleYSeq          tween.Sequence
}

func (intro *LevelIntro) Init(levelTitle, mapNumber string) {
	*intro = LevelIntro{}

	if len(levelTitle) == 0 || len(mapNumber) == 0 {
		intro.timer = math2.Inf32()
		return
	}

	// Black background
	intro.blackBackground = ui.NewColorBox(ui.Transform{
		Anchor: ui.Ratios{0.5, 0.5},
		Origin: ui.Ratios{0.5, 0.5},
		Size:   mgl32.Vec2{float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight)},
		Depth:  8.9,
	}, color.Black)

	// Star background
	starMesh, _ := cache.GetMesh("assets/models/star_transition.obj")
	intro.starBackground = ui.Element{
		BgMesh:  starMesh,
		BgColor: maybe.Some(color.Black),
	}
	intro.starBackground.SetTransform(ui.Transform{
		Anchor: ui.Ratios{0.5, 0.5},
		Origin: ui.Ratios{0.5, 0.5},
		Size:   mgl32.Vec2{float32(settings.Current.WindowWidth), float32(settings.Current.WindowHeight)},
		Depth:  8.9,
	})

	// Top banner
	intro.bannerTop = ui.NewText(ui.Transform{
		Position: mgl32.Vec2{0.0, 80.0},
		Size:     mgl32.Vec2{float32(settings.Current.WindowWidth), 96.0},
		Depth:    9.1,
	}, levelTitle, ui.TextConfig{
		Color: maybe.Some(color.White),
		Align: ui.TextAlignCenterH | ui.TextAlignCenterV,
		Scale: maybe.Some[float32](3.0),
	})
	intro.bannerTop.BgColor = maybe.Some(color.Blue)
	intro.bannerTop.BgMesh = cache.QuadMesh

	// Bottom banner
	intro.bannerBottom = ui.NewText(ui.Transform{
		Position: mgl32.Vec2{-96.0, settings.UIHeight() - 224.0},
		Size:     mgl32.Vec2{448.0, 96.0},
		Depth:    9.1,
		Shear:    mgl32.Vec2{1.0, 0.0},
	}, mapNumber, ui.TextConfig{
		Align: ui.TextAlignCenterH | ui.TextAlignCenterV,
		Scale: maybe.Some[float32](3),
	})
	intro.bannerBottom.BgMesh = cache.QuadMesh
	intro.bannerBottom.BgColor = maybe.Some(color.Blue)

	// Sickle
	sickleTex := cache.GetTexture("assets/textures/ui/intro_sickle.png")
	sickleSize := sickleTex.Rect().SizeVec().Mul(SpriteScale())
	intro.sickle = ui.NewBox(ui.Transform{
		Origin:   ui.Ratios{0.5, 0.5},
		Position: mgl32.Vec2{settings.UIWidth() + sickleSize[0] + 8.0, intro.bannerTop.Y() + intro.bannerTop.Height()/2.0},
		Size:     sickleSize,
		Depth:    9.3,
	}, sickleTex)
	intro.sickleXSeq = tween.Sequence{Tweens: []tween.Data{
		{StartValue: intro.sickle.X(), EndValue: tween.Infer, Duration: 0.1},
		{StartValue: tween.Infer, EndValue: -sickleSize[0] - 8.0, Duration: 0.5},
		{StartValue: tween.Infer, EndValue: tween.Infer, Duration: 0.5},
		{StartValue: tween.Infer, EndValue: intro.sickle.X(), Duration: 0.5},
		{StartValue: tween.Infer, EndValue: tween.Infer, Duration: 1.0},
	}}
	intro.sickleYSeq = tween.Sequence{Tweens: []tween.Data{
		{StartValue: intro.sickle.Y(), EndValue: tween.Infer, Duration: 1.0},
		{StartValue: intro.bannerBottom.Y() + intro.bannerBottom.Height()/2.0, EndValue: tween.Infer, Duration: 1.0},
	}}

	// Eyes
	eyesTex := cache.GetTexture("assets/textures/ui/intro_eyes.png")
	eyesAnim := eyesTex.GetDefaultAnimation()
	intro.eyes = ui.NewBox(ui.Transform{
		Origin: ui.Ratios{0.5, 0.5},
		Anchor: ui.Ratios{0.5, 0.5},
		Depth:  9.3,
		Size:   eyesAnim.Frames[0].Rect.SizeVec().Mul(SpriteScale()),
	}, eyesTex)
	intro.eyes.AnimPlayer = comps.NewAnimationPlayer(eyesAnim, true)
}

func (intro *LevelIntro) Done() bool {
	return intro.timer >= LevelIntroTime
}

// Returns number of seconds left before the intro ends.
func (intro *LevelIntro) TimeLeft() float32 {
	return LevelIntroTime - intro.timer
}

func (intro *LevelIntro) Layout(queue *ui.RenderQueue, deltaTime float32) {
	intro.timer += deltaTime

	intro.eyes.AnimPlayer.Update(deltaTime)
	queue.Add(&intro.eyes)

	// Move sickle
	intro.sickle.Rotate(math2.Radians(deltaTime * math.Pi * 12.0))
	sickleXRes := intro.sickleXSeq.Update(deltaTime)
	intro.sickle.SetX(sickleXRes.Value)
	sickleYRes := intro.sickleYSeq.Update(deltaTime)
	intro.sickle.SetY(sickleYRes.Value)
	if sickleXRes.TweenDone {
		switch sickleXRes.SequenceIndex {
		case 0:
			intro.voice = cache.GetSfx("assets/sounds/ui/intro_whoosh1.wav").Play()
		case 2:
			intro.voice = cache.GetSfx("assets/sounds/ui/intro_whoosh2.wav").Play()
		}
	}
	queue.Add(&intro.sickle)

	// Occlude banners along with sickle
	switch sickleXRes.SequenceIndex {
	case 1:
		intro.bannerTop.Scissor = math2.Rect{
			X: max(0, intro.sickle.X()), Y: 0,
			Width: max(1, settings.UIWidth()-intro.sickle.X()), Height: settings.UIHeight(),
		}
		queue.Add(&intro.bannerTop)
	case 2, 3, 4:
		intro.bannerTop.Scissor = math2.Rect{}
		intro.bannerBottom.Scissor = math2.Rect{X: 0.0, Y: 0, Width: max(0, intro.sickle.X()), Height: settings.UIHeight()}
		queue.Add(&intro.bannerTop, &intro.bannerBottom)
	default:
		intro.bannerTop.Scissor, intro.bannerBottom.Scissor = math2.Rect{}, math2.Rect{}
	}

	if sickleXRes.SequenceDone {
		// Move banners off screen at the end
		delta := mgl32.Vec2{deltaTime * settings.UIWidth() * 2.0, 0}
		intro.bannerTop.Translate(delta)
		intro.bannerBottom.Translate(delta.Mul(-1))
	}

	if intro.timer < 2.0 {
		// Black background
		queue.Add(&intro.blackBackground)
	} else {
		// Star background
		intro.starBackground.SetSize(intro.starBackground.Size().Mul(1.0 + (deltaTime * 25.0)))
		queue.Add(&intro.starBackground)
	}
}
