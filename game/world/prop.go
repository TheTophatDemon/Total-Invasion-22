package world

import (
	"fmt"
	"strings"

	"tophatdemon.com/total-invasion-ii/engine/assets/audio"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

type PropType uint8

const (
	PROP_TYPE_GENERIC PropType = iota
	PROP_TYPE_GEOFFREY
)

// An unmoving object in the game world used as decoration
type Prop struct {
	SpriteRender comps.SpriteRender
	AnimPlayer   comps.AnimationPlayer
	body         comps.Body
	world        *World
	propType     PropType
	voice        audio.VoiceId
}

var _ comps.HasBody = (*Prop)(nil)
var _ Usable = (*Prop)(nil)

func (p *Prop) Body() *comps.Body {
	return &p.body
}

func (p *Prop) OnUse(player *Player) {
	switch p.propType {
	case PROP_TYPE_GEOFFREY:
		p.world.ShowMessage(settings.Localize("geoffrey"), 2.0, 10, color.Red)
	}
}

func SpawnPropFromTE3(st *scene.Storage[Prop], world *World, ent te3.Ent) (id scene.Id[*Prop], prop *Prop, err error) {
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

	id, prop, err = st.New()
	if err != nil {
		return
	}

	prop.world = world

	sprite := cache.GetTexture(texturePath)
	prop.SpriteRender = comps.NewSpriteRender(sprite)

	anims := sprite.GetAnimationNames()
	if len(anims) > 0 {
		anim, _ := sprite.GetAnimation(anims[0])
		prop.AnimPlayer = comps.NewAnimationPlayer(anim, true)
	}

	radius, err := ent.FloatProperty("radius")
	if err != nil {
		radius = ent.Radius
		err = nil
	}

	prop.body = comps.Body{
		Transform: comps.TransformFromTE3Ent(ent, true, true),
		Shape:     collision.NewSphere(radius),
		Layer:     COL_LAYER_MAP,
		Filter:    COL_LAYER_NONE,
	}

	switch strings.ToLower(ent.Properties["prop"]) {
	case "geoffrey":
		prop.propType = PROP_TYPE_GEOFFREY
	}

	return
}

func (prop *Prop) Update(deltaTime float32) {
	prop.AnimPlayer.Update(deltaTime)
	if prop.voice.IsValid() {
		prop.voice.Attenuate(prop.Body().Transform.Position(), prop.world.ListenerTransform())
	}
}

func (prop *Prop) Render(context *render.Context) {
	prop.SpriteRender.Render(&prop.body.Transform, &prop.AnimPlayer, context, prop.body.Transform.Yaw())
}
