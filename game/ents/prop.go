package ents

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/world/comps"
)

// An unmoving object in the game world used as decoration
type Prop struct {
	SpriteRender comps.SpriteRender
	AnimPlayer   comps.AnimationPlayer
	body         comps.Body
	world        WorldOps
}

var _ comps.HasBody = (*Prop)(nil)
var _ Usable = (*Prop)(nil)

func (p *Prop) Body() *comps.Body {
	return &p.body
}

func (p *Prop) OnUse(player *Player) {
	fmt.Println("Woop!")
}

func NewPropFromTE3(ent te3.Ent, world WorldOps) (prop Prop, err error) {
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

	prop.world = world

	sprite := cache.GetTexture(texturePath)
	prop.SpriteRender = comps.NewSpriteRender(sprite)

	anims := sprite.GetAnimationNames()
	if len(anims) > 0 {
		anim, _ := sprite.GetAnimation(anims[0])
		prop.AnimPlayer = comps.NewAnimationPlayer(anim, true)
	}

	prop.body = comps.Body{
		Transform: ent.Transform(false),
		Shape:     collision.NewSphere(ent.Radius),
		Pushiness: 10_000,
	}

	return
}

func (p *Prop) Update(deltaTime float32) {
	p.AnimPlayer.Update(deltaTime)
	p.body.Update(deltaTime)
}

func (p *Prop) Render(context *render.Context) {
	p.SpriteRender.Render(&p.body.Transform, &p.AnimPlayer, context, p.body.Transform.Yaw())
}
