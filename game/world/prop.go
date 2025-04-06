package world

import (
	"fmt"
	"strings"

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

type PropType uint8

const (
	PROP_TYPE_GENERIC PropType = iota
	PROP_TYPE_GEOFFREY
	PROP_TYPE_FIRE
)

const (
	GEOFFREY_SAFETY_RADIUS = 2.5
	GEOFFREY_COL_LAYERS    = COL_LAYER_MAP | COL_LAYER_NPCS
	SFX_HONK               = "assets/sounds/honk.wav"
	GEOFFREY_ANIM_IDLE     = "idle"
	GEOFFREY_ANIM_VANISH   = "vanish"
)

// A (generally) unmoving object in the game world used as decoration
type Prop struct {
	id           scene.Id[*Prop]
	SpriteRender comps.SpriteRender
	AnimPlayer   comps.AnimationPlayer
	debugShape   DebugShape
	body         comps.Body
	world        *World
	propType     PropType
	voice        tdaudio.VoiceId
	isSeen       bool
	radius       float32
}

var _ comps.HasBody = (*Prop)(nil)
var _ Usable = (*Prop)(nil)

func (p *Prop) Body() *comps.Body {
	return &p.body
}

func (p *Prop) OnUse(player *Player) {
	switch p.propType {
	case PROP_TYPE_GEOFFREY:
		p.world.Hud.ShowMessage(settings.Localize("geoffrey"), 2.0, 10, color.Red)
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

	sprite := cache.GetTexture(texturePath)
	prop.SpriteRender = comps.NewSpriteRender(sprite)

	anim := sprite.GetDefaultAnimation()
	if anim.Frames != nil {
		prop.AnimPlayer = comps.NewAnimationPlayer(anim, true)
	}

	prop.radius, err = ent.FloatProperty("radius")
	if err != nil {
		prop.radius = ent.Radius
		err = nil
	}

	// mesh, err := cache.GetMesh("assets/models/shapes/cylinder.obj")
	// if err != nil {
	// 	log.Printf("Error loading cylinder mesh for prop: %v.\n", err)
	// }

	// shapeMesh := collision.NewMesh(mesh, prop.radius)
	prop.body = comps.Body{
		Transform: comps.TransformFromTE3Ent(ent, true, true),
		// Shape:     shapeMesh,
		Shape:  collision.NewSphere(prop.radius),
		Layer:  COL_LAYER_MAP,
		Filter: COL_LAYER_NONE,
	}

	// wireMesh := collision.WireMeshFromMeshCollisionShape(&shapeMesh, color.Red)

	// prop.debugShape = DebugShape{
	// 	MeshRender: comps.NewMeshRender(wireMesh, shaders.DebugShader, nil),
	// 	Transform:  prop.body.Transform,
	// 	TimeLeft:   math2.Inf32(),
	// }

	switch strings.ToLower(ent.Properties["prop"]) {
	case "geoffrey":
		prop.propType = PROP_TYPE_GEOFFREY
		prop.body.Layer = GEOFFREY_COL_LAYERS
	case "fire":
		prop.propType = PROP_TYPE_FIRE
		prop.body.Layer = COL_LAYER_NONE
		SpawnKillzone(world, prop.body.Transform.Position(), 0.5, 25.0)
	}

	return
}

func (prop *Prop) Update(deltaTime float32) {
	prop.AnimPlayer.Update(deltaTime)
	prop.voice.SetPositionV(prop.Body().Transform.Position())

	switch prop.propType {
	case PROP_TYPE_GEOFFREY:
		vanishAnim, _ := prop.SpriteRender.Texture().GetAnimation(GEOFFREY_ANIM_VANISH)
		if !prop.isSeen && len(prop.world.BodiesInSphere(prop.body.Transform.Position(), prop.radius, prop)) == 0 {
			// Make Geoffrey re-appear when nobody is looking.
			if prop.AnimPlayer.IsPlayingAnim(vanishAnim) && prop.AnimPlayer.IsAtEnd() {
				idleAnim, _ := prop.SpriteRender.Texture().GetAnimation(GEOFFREY_ANIM_IDLE)
				prop.AnimPlayer.PlayNewAnim(idleAnim)
				prop.body.Layer = GEOFFREY_COL_LAYERS
			}
		} else if !prop.AnimPlayer.IsPlayingAnim(vanishAnim) {
			// Check for incoming projectiles and trigger the disappearing animation.
			if prop.world.AnyProjectilesInSphere(prop.body.Transform.Position(), GEOFFREY_SAFETY_RADIUS) {
				prop.AnimPlayer.PlayNewAnim(vanishAnim)
				prop.body.Layer = 0
				cache.GetSfx(SFX_HONK).PlayAttenuatedV(prop.body.Transform.Position())
			}
		}
	}
}

func (prop *Prop) Render(context *render.Context) {
	prop.isSeen = prop.SpriteRender.Render(&prop.body.Transform, &prop.AnimPlayer, context, prop.body.Transform.Yaw())
	prop.debugShape.Render(context)
}
