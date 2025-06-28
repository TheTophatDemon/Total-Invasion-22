package hud

import (
	"math"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const LEVEL_INTRO_TIME = 3.0 // Time after which the level intro ends.

type LevelIntro struct {
	timer                   float32
	voice                   tdaudio.VoiceId
	starBackground          ui.Box
	sweepTop, sweepBottom   ui.Box // Black boxes covering up the level titles until the sickle flies past them.
	bannerTop, bannerBottom ui.Box
	titleText, episodeText  ui.Text
	sickle, eyes            ui.Box
}

func (intro *LevelIntro) Init(levelTitle, mapNumber string) {
	*intro = LevelIntro{}

	if len(levelTitle) == 0 || len(mapNumber) == 0 {
		intro.timer = math2.Inf32()
		return
	}

	// Star background
	starMesh, _ := cache.GetMesh("assets/models/star_transition.obj")
	intro.starBackground = ui.Box{
		Transform: ui.Transform{
			Dest: math2.Rect{
				Y:      -settings.UIHeight() * 0.25,
				Width:  settings.UIWidth(),
				Height: settings.UIHeight() * 1.5,
			},
			Depth: 8.9,
			Scale: 0.5,
		},
		Color: color.Black,
		Mesh:  starMesh,
	}

	// Top banner
	intro.bannerTop = ui.Box{
		Color: color.Blue,
		Transform: ui.Transform{
			Dest:  math2.Rect{X: 0.0, Y: 64.0, Width: settings.UIWidth(), Height: 96.0},
			Depth: 9.0,
		},
	}

	intro.titleText = ui.Text{
		Color: color.White,
		Settings: ui.TextSettings{
			Text:      levelTitle,
			Alignment: ui.TEXT_ALIGN_CENTER,
			Font:      cache.DefaultFont,
		},
		Transform: ui.Transform{
			Dest: math2.Rect{
				Y:      80.0,
				Width:  settings.UIWidth(),
				Height: 64.0,
			},
			Scale: 3.0,
			Depth: 9.1,
		},
	}

	intro.sweepTop = ui.Box{
		Color: color.Black,
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      0.0,
				Y:      63.0,
				Width:  settings.UIWidth(),
				Height: 98.0,
			},
			Depth: 9.2,
		},
		// Set src? Matrix?
	}

	// Bottom banner
	intro.bannerBottom = ui.Box{
		Color: color.Blue,
		Transform: ui.Transform{
			Dest:  math2.Rect{X: -48.0, Y: settings.UIHeight() - 224.0, Width: 352.0, Height: 96.0},
			Depth: 9.0,
			Shear: 1.0,
		},
	}

	intro.episodeText = ui.Text{
		Color: color.White,
		Settings: ui.TextSettings{
			Text:      mapNumber,
			Alignment: ui.TEXT_ALIGN_LEFT,
			Font:      cache.DefaultFont,
		},
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      32.0,
				Y:      intro.bannerBottom.DestPosition()[1] + 8.0,
				Width:  256.0,
				Height: intro.bannerBottom.Dest.Height,
			},
			Scale: 3.0,
			Depth: 9.1,
		},
	}

	intro.sweepBottom = ui.Box{
		Color: color.Black,
		Transform: ui.Transform{
			Dest: math2.Rect{
				X:      0.0,
				Y:      intro.bannerBottom.Dest.Y - 1.0,
				Width:  intro.bannerBottom.Dest.Width + 1.0,
				Height: intro.bannerBottom.Dest.Height + 2.0,
			},
			Depth: 9.2,
		},
	}

	sickleTex := cache.GetTexture("assets/textures/ui/intro_sickle.png")
	intro.sickle = ui.NewBoxFull(math2.Rect{
		X:      settings.UIWidth() + 8.0,
		Y:      intro.sweepTop.Dest.Y - float32(sickleTex.Height())/2.0,
		Width:  float32(sickleTex.Width()) * SpriteScale(),
		Height: float32(sickleTex.Height()) * SpriteScale(),
	}, sickleTex, color.White)
	intro.sickle.Depth = 9.3

	eyesTex := cache.GetTexture("assets/textures/ui/intro_eyes.png")
	intro.eyes = ui.Box{
		Color:      color.White,
		AnimPlayer: comps.NewAnimationPlayer(eyesTex.GetDefaultAnimation(), true),
		Texture:    eyesTex,
	}
	eyesWidth := intro.eyes.AnimPlayer.Frame().Rect.Width * SpriteScale()
	eyesHeight := intro.eyes.AnimPlayer.Frame().Rect.Height * SpriteScale()
	intro.eyes.Transform = ui.Transform{
		Dest: math2.Rect{
			X:      settings.UIWidth()/2.0 - eyesWidth/2.0,
			Y:      settings.UIHeight()/2.0 - eyesHeight/2.0,
			Width:  eyesWidth,
			Height: eyesHeight,
		},
		Depth: 9.3,
	}
}

func (intro *LevelIntro) Done() bool {
	return intro.timer >= LEVEL_INTRO_TIME
}

// Returns number of seconds left before the intro ends.
func (intro *LevelIntro) TimeLeft() float32 {
	return LEVEL_INTRO_TIME - intro.timer
}

func (intro *LevelIntro) Layout(queue *ui.RenderQueue, deltaTime float32) {
	intro.timer += deltaTime

	intro.eyes.Update(deltaTime)
	queue.Add(&intro.eyes)

	// Move sickle
	sickleSpeed := deltaTime * settings.UIWidth() * 3.0
	intro.sickle.Rotation += deltaTime * math.Pi * 16.0
	switch {
	case intro.timer < 0.5:
		// Wait
		if !intro.voice.IsValid() {
			intro.voice = cache.GetSfx("assets/sounds/ui/intro_whoosh1.wav").Play()
		}
	case intro.timer < 1.0:
		intro.voice = tdaudio.VoiceId{}
		if intro.sickle.Dest.X < -float32(intro.sickle.Texture.Width())*SpriteScale() {
			intro.sickle.Dest.Y = intro.sweepBottom.Dest.Y - float32(intro.sickle.Texture.Height()/2)*SpriteScale()
		} else {
			intro.sickle.Dest.X -= sickleSpeed
		}
	case intro.timer < 3.5:
		if !intro.voice.IsValid() {
			intro.voice = cache.GetSfx("assets/sounds/ui/intro_whoosh2.wav").Play()
		}
		intro.sickle.Dest.X += sickleSpeed
	}
	queue.Add(&intro.sickle)

	if intro.timer < 2.0 {
		// Black background
		queue.Add(&ui.Box{
			Color: color.Black,
			Transform: ui.Transform{
				Dest:  math2.Rect{Width: settings.UIWidth(), Height: settings.UIHeight()},
				Depth: 8.9,
			},
		})
	} else {
		// Star background
		intro.starBackground.Scale += deltaTime * 200.0
		queue.Add(&intro.starBackground)
	}

	// Move top sweep
	if intro.timer > 0.5 {
		intro.sweepTop.Dest.Width -= deltaTime * settings.UIWidth() * 3.0
	}
	queue.Add(&intro.sweepTop)

	// Move bottom sweep
	if intro.timer > 1.0 {
		delta := deltaTime * 2048.0
		intro.sweepBottom.Dest.Width = max(0, intro.sweepBottom.Dest.Width-delta)
		intro.sweepBottom.Dest.X += delta
	}
	queue.Add(&intro.sweepBottom)

	// Move titles and banners
	if intro.timer > 2.5 {
		delta := deltaTime * settings.UIWidth() * 2.0
		intro.bannerTop.Dest.X += delta
		intro.titleText.Dest.X += delta
		intro.bannerBottom.Dest.X -= delta
		intro.episodeText.Dest.X -= delta
	}
	queue.Add(&intro.bannerTop, &intro.titleText, &intro.bannerBottom, &intro.episodeText)
}
