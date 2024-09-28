package world

import (
	"fmt"

	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
)

type Item struct {
	body         comps.Body
	spriteRender comps.SpriteRender
	animPlayer   comps.AnimationPlayer
	flashColor   color.Color
	pickupSound  tdaudio.SoundId
	healAmount   float32
	world        *World
	id           scene.Id[*Item]
}

var _ comps.HasBody = (*Item)(nil)

func SpawnItemFromTE3(world *World, ent te3.Ent) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = world.Items.New()
	if err != nil {
		return
	}

	itemType, isItem := ent.Properties["item"]
	if !isItem {
		id.Remove()
		return scene.Id[*Item]{}, nil, fmt.Errorf("item is missing 'item' property")
	}

	*item = Item{
		body: comps.Body{
			Transform: comps.TransformFromTE3Ent(ent, false, false),
			Shape:     collision.NewSphere(0.5),
			Layer:     0,
			Filter:    COL_LAYER_PLAYERS,
		},
		flashColor:  color.White.WithAlpha(0.75),
		pickupSound: cache.GetSfx("assets/sounds/pickup.wav"),
		world:       world,
		id:          id,
	}

	var textureName string
	switch itemType {
	case "medkit":
		item.healAmount = 50.0
		item.body.Transform.SetScaleUniform(0.375)
		textureName = "assets/textures/sprites/medkit.png"
	case "stimpack":
		item.healAmount = 10.0
		item.body.Transform.SetScaleUniform(0.25)
		textureName = "assets/textures/sprites/stimpack.png"
	case "cartonofeggs", "egg_carton":
		//TODO: Ammo types...
		item.body.Transform.SetScaleUniform(0.5)
		textureName = "assets/textures/sprites/egg_carton.png"
	default:
		id.Remove()
		return scene.Id[*Item]{}, nil, fmt.Errorf("item type '%v' is not implemented yet", itemType)
	}

	// Put the item on the bottom side of its grid cel.
	item.body.Transform.Translate(0.0, item.body.Transform.Scale().Y()-1.0, 0.0)

	texture := cache.GetTexture(textureName)
	item.spriteRender = comps.NewSpriteRender(texture)
	if texture.AnimationCount() > 0 {
		item.animPlayer = comps.NewAnimationPlayer(texture.GetDefaultAnimation(), true)
	}
	item.body.OnIntersect = item.onIntersect

	return
}

func (item *Item) Body() *comps.Body {
	return &item.body
}

func (item *Item) Update(deltaTime float32) {
	item.animPlayer.Update(deltaTime)
}

func (item *Item) Render(context *render.Context) {
	item.spriteRender.Render(&item.body.Transform, &item.animPlayer, context, 0.0)
}

func (item *Item) onIntersect(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	player, isPlayer := otherEnt.(*Player)
	if !otherEnt.Body().OnLayer(item.body.Filter) || !isPlayer {
		return
	}

	player.actor.Health += item.healAmount

	item.pickupSound.PlayAttenuatedV(item.body.Transform.Position())
	item.world.Hud.FlashScreen(item.flashColor, 1.5)
	item.world.QueueRemoval(item.id.Handle)
}
