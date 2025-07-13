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
	world         *World
	voice         tdaudio.VoiceId
	isSeen        bool
	radius        float32
	stareTimer    float32
	entProperties map[string]string
	updateFunc    func(deltaTime float32)
	useFunc       func(player *Player)
}

func (prop *Prop) Body() *comps.Body {
	return &prop.body
}

func (prop *Prop) OnUse(player *Player) {
	if prop.useFunc != nil {
		prop.useFunc(player)
	}
}

func SpawnPropFromTE3(world *World, ent te3.Ent) (id scene.Id[*Prop], prop *Prop, err error) {
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

	id, prop, err = world.Props.New()
	if err != nil {
		return
	}

	prop.id = id
	prop.world = world
	prop.entProperties = ent.Properties

	sprite := cache.GetTexture(texturePath)

	anim := sprite.GetDefaultAnimation()
	if anim.Frames != nil {
		prop.AnimPlayer = comps.NewAnimationPlayer(anim, true)
	}

	prop.radius, err = ent.FloatProperty("radius")
	if err != nil {
		prop.radius = ent.Radius
		err = nil
	}

	prop.body = comps.Body{
		Transform: comps.TransformFromTE3Ent(ent, true, true),
		Shape:     collision.NewCylinder(prop.radius, 2.0),
		Layer:     ColLayerMap,
		Filter:    ColLayerNone,
	}

	if prop.radius == 0 {
		prop.body.Layer = ColLayerNone
	}

	colr := color.White
	additive := false

	switch strings.ToLower(ent.Properties["prop"]) {
	case "geoffrey":
		prop.updateFunc = prop.geoffreyUpdate
		prop.useFunc = prop.geoffreyUse
		prop.body.Layer = ColLayerMap | ColLayerNPCs
	case "eyeball":
		prop.updateFunc = prop.eyeballUpdate
		prop.body.Layer = ColLayerMap | ColLayerNPCs
	case "fire":
		prop.body.Layer = ColLayerInvisible
		colr = color.Color{R: 1.0, G: 1.0, B: 1.0, A: 0.5}
		additive = true
		prop.body.Transform.SetScale(1.0, 1.25, 1.0)
		prop.body.Transform.Translate(0.0, 0.25, 0.0)
	}

	prop.SpriteRender = comps.NewSpriteRenderWithColor(sprite, colr)
	prop.SpriteRender.AdditiveBlending = additive

	return
}

func (prop *Prop) Update(deltaTime float32) {
	prop.AnimPlayer.Update(deltaTime)
	particlesTransform := prop.Body().Transform
	particlesTransform.TranslateV(mgl32.Vec3{0.0, 0.0, 0.0})
	prop.voice.SetPositionV(prop.Body().Transform.Position())

	if prop.updateFunc != nil {
		prop.updateFunc(deltaTime)
	}
}

func (prop *Prop) DistanceFromScreen(context *render.Context) float32 {
	return mgl32.TransformCoordinate(
		mgl32.Vec3{0.0, 0.0, 1.0},
		context.ViewInverse.Mul4(prop.body.Transform.Matrix()),
	)[2]
}

func (prop *Prop) Render(context *render.Context) {
	if prop.SpriteRender.DiffuseColor.A < 1.0 && !context.DrawingTranslucent {
		context.EnqueueTranslucentRender(prop)
	} else {
		prop.isSeen = prop.SpriteRender.Render(&prop.body.Transform, &prop.AnimPlayer, context, prop.body.Transform.Yaw())
	}
}

/************
 * GEOFFREY *
 ************/

func (prop *Prop) geoffreyUse(player *Player) {
	prop.world.Hud.ShowMessage(settings.Localize("geoffrey"), 2.0, 10, color.Red)
}

func (prop *Prop) geoffreyUpdate(deltaTime float32) {
	vanishAnim, _ := prop.SpriteRender.Texture().GetAnimation("vanish")
	bodiesIter := prop.world.IterBodiesInSphere(prop.body.Transform.Position(), prop.radius, prop)
	if !prop.isSeen && !bodiesIter.HasNext() {
		// Make Geoffrey re-appear when nobody is looking.
		if prop.AnimPlayer.IsPlayingAnim(vanishAnim) && prop.AnimPlayer.IsAtEnd() {
			idleAnim, _ := prop.SpriteRender.Texture().GetAnimation("idle")
			prop.AnimPlayer.PlayNewAnim(idleAnim)
			prop.body.Layer = ColLayerMap | ColLayerNPCs
		}
	} else if !prop.AnimPlayer.IsPlayingAnim(vanishAnim) {
		// Check for incoming projectiles and trigger the disappearing animation.
		projsIter := prop.world.IterProjectilesInSphere(prop.body.Transform.Position(), projectileSafetyRadius, nil)
		if projsIter.HasNext() {
			prop.AnimPlayer.PlayNewAnim(vanishAnim)
			prop.body.Layer = 0
			cache.GetSfx("assets/sounds/honk.wav").PlayAttenuatedV(prop.body.Transform.Position())
		}
	}
}

/***********
 * EYEBALL *
 ***********/

func (prop *Prop) eyeballUpdate(deltaTime float32) {
	idleAnim, _ := prop.SpriteRender.Texture().GetAnimation("idle")
	openAnim, _ := prop.SpriteRender.Texture().GetAnimation("open")
	stareAnim, _ := prop.SpriteRender.Texture().GetAnimation("stare")
	eyeContact := false
	projsIter := prop.world.IterProjectilesInSphere(prop.body.Transform.Position(), projectileSafetyRadius, nil)
	if !projsIter.HasNext() {
		if camera, ok := prop.world.CurrentCamera.Get(); ok && camera.Position() != prop.body.Transform.Position() {
			toCamera := camera.Position().Sub(prop.body.Transform.Position()).Normalize()
			if camera.Forward().Dot(toCamera) < -0.95 {
				res, handle := prop.world.Raycast(prop.body.Transform.Position(), toCamera, ColLayerMap|ColLayerNPCs|ColLayerPlayers, 15.0, prop)
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
			prop.AnimPlayer.PlayAnimSequence(openAnim, idleAnim)
		}
	} else if prop.stareTimer > 1.0 && prop.stareTimer < 1.5 {
		prop.world.Hud.ShowMessage(settings.Localize(prop.entProperties["messageKey"]), 1.0, 50, color.Magenta)
	}
}
