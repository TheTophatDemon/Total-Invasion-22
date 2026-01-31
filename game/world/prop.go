package world

import (
	"fmt"
	"strings"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

const projectileSafetyRadius = 2.5

// A (generally) unmoving object in the game world used as decoration
type Prop struct {
	id            scene.Id[*Prop]
	SpriteRender  comps.SpriteRender
	AnimPlayer    comps.AnimationPlayer
	body          comps.Body
	voice         tdaudio.VoiceId
	isSeen        bool
	radius        float32
	stareTimer    float32
	entProperties map[string]string
	updateFunc    func(deltaTime float32)
	useFunc       func(player *Player)
	yaw           float32
}

func (prop *Prop) Body() *comps.Body {
	return &prop.body
}

func (prop *Prop) OnUse(player *Player) {
	if prop.useFunc != nil {
		prop.useFunc(player)
	}
}

func SpawnPropFromTE3(ent te3.Ent) (id scene.Id[*Prop], prop *Prop, err error) {
	if ent.Display != te3.ENT_DISPLAY_SPHERE && ent.Display != te3.ENT_DISPLAY_SPRITE {
		err = fmt.Errorf("te3 ent display mode should be 'sprite' or 'sphere'")
		return
	}

	texturePath, ok := ent.Properties["texture"]
	if !ok && len(ent.Texture) == 0 {
		err = fmt.Errorf("prop is missing texture")
		return
	} else if !ok {
		texturePath = ent.Texture
	}

	id, prop, err = gWorld.Props.New()
	if err != nil {
		return
	}

	prop.id = id
	prop.entProperties = ent.Properties
	prop.yaw = ent.AnglesInRadians()[1]

	sprite := cache.GetTexture(texturePath)

	anim := sprite.GetDefaultAnimation()
	if anim.Frames != nil {
		prop.AnimPlayer = comps.NewAnimationPlayer(anim, true)
	}

	prop.radius, err = ent.FloatProperty("radius")
	if err != nil {
		prop.radius = 0.5
		err = nil
	}

	prop.body = comps.Body{
		Position: mgl32.Vec3(ent.Position).Add(mgl32.Vec3{0.0, ent.Radius - 1.0, 0.0}),
		Shape:    collision.NewBoxShape(prop.radius, 2.0, prop.radius),
		Layers:   ColLayerMap,
	}

	if prop.radius == 0 {
		prop.body.Layers = ColLayerNone
	}

	spriteScale := mgl32.Vec2{ent.Radius, ent.Radius}

	switch strings.ToLower(ent.Properties["prop"]) {
	case "geoffrey":
		prop.updateFunc = prop.geoffreyUpdate
		prop.useFunc = prop.geoffreyUse
		prop.body.Layers = ColLayerMap | ColLayerNPCs | ColLayerUsable
	case "eyeball":
		prop.updateFunc = prop.eyeballUpdate
		prop.useFunc = prop.eyeballUse
		prop.body.Layers = ColLayerMap | ColLayerNPCs | ColLayerUsable
	case "fire":
		prop.body.Layers = ColLayerInvisible
	}

	prop.SpriteRender = comps.NewSpriteRender(sprite, nil, &spriteScale)

	return
}

func (prop *Prop) Update(deltaTime float32) {
	prop.AnimPlayer.Update(deltaTime)
	prop.voice.SetPositionV(prop.Body().Position)

	if prop.updateFunc != nil {
		prop.updateFunc(deltaTime)
	}
}

func (prop *Prop) Render(context *render.Context) {
	prop.isSeen = prop.SpriteRender.Render(prop.body.Position, &prop.AnimPlayer, context, prop.yaw)
}

/************
 * GEOFFREY *
 ************/

func (prop *Prop) geoffreyUse(player *Player) {
	gWorld.Hud.ShowMessage(settings.Localize("geoffrey"), 2.0, 10, color.Red)
}

func (prop *Prop) geoffreyUpdate(deltaTime float32) {
	vanishAnim, _ := prop.SpriteRender.Texture().GetAnimation("vanish")
	bodiesIter := gWorld.IterBodiesInSphere(prop.body.Position, prop.radius, prop)
	if !prop.isSeen && !bodiesIter.HasNext() {
		// Make Geoffrey re-appear when nobody is looking.
		if prop.AnimPlayer.IsPlayingAnim(vanishAnim) && prop.AnimPlayer.IsAtEnd() {
			idleAnim, _ := prop.SpriteRender.Texture().GetAnimation("idle")
			prop.AnimPlayer.PlayNewAnim(idleAnim)
			prop.body.Layers = ColLayerMap | ColLayerNPCs
		}
	} else if !prop.AnimPlayer.IsPlayingAnim(vanishAnim) {
		// Check for incoming projectiles and trigger the disappearing animation.
		projsIter := gWorld.IterProjectilesInSphere(prop.body.Position, projectileSafetyRadius, nil)
		if projsIter.HasNext() {
			prop.AnimPlayer.PlayNewAnim(vanishAnim)
			prop.body.Layers = 0
			cache.GetSfx("assets/sounds/honk.wav").PlayAttenuatedV(prop.body.Position)
		}
	}
}

/***********
 * EYEBALL *
 ***********/

func (prop *Prop) eyeballUse(player *Player) {
	_ = player
	gWorld.Hud.ShowMessage(settings.Localize(prop.entProperties["messageKey"]), 1.0, 50, color.Magenta)
}

func (prop *Prop) eyeballUpdate(deltaTime float32) {
	idleAnim, _ := prop.SpriteRender.Texture().GetAnimation("idle")
	openAnim, _ := prop.SpriteRender.Texture().GetAnimation("open")
	closeAnim, _ := prop.SpriteRender.Texture().GetAnimation("close")
	stareAnim, _ := prop.SpriteRender.Texture().GetAnimation("stare")
	eyeContact := false
	projsIter := gWorld.IterProjectilesInSphere(prop.body.Position, projectileSafetyRadius, nil)
	if !projsIter.HasNext() {
		if camera, ok := gWorld.CurrentCamera.Get(); ok && camera.Position() != prop.body.Position {
			toCamera := camera.Position().Sub(prop.body.Position).Normalize()
			if camera.Forward().Dot(toCamera) < -0.95 {
				res, handle := gWorld.Raycast(prop.body.Position, toCamera, ColLayerMap|ColLayerNPCs|ColLayerPlayers, 7.5, prop.Body())
				if _, isPlayer := scene.Get[*Player](handle); res.Hit && isPlayer {
					prop.stareTimer += deltaTime
					eyeContact = true
					if prop.AnimPlayer.IsPlayingAnim(idleAnim) {
						prop.AnimPlayer.PlayAnimSequence(openAnim, stareAnim)
					}
				}
			}
		}
	}
	if !eyeContact {
		prop.stareTimer = 0.0
		if prop.AnimPlayer.IsPlayingAnim(stareAnim) {
			prop.AnimPlayer.PlayAnimSequence(closeAnim, idleAnim)
		}
	} else if prop.stareTimer > 1.0 && prop.stareTimer < 1.5 {
		prop.eyeballUse(nil)
	}
}
