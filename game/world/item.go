package world

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/math2/collision"
	"tophatdemon.com/total-invasion-ii/engine/render"
	"tophatdemon.com/total-invasion-ii/engine/scene"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
)

type Item struct {
	body          comps.Body
	spriteRender  comps.SpriteRender
	animPlayer    comps.AnimationPlayer
	flashColor    color.Color
	pickupSound   tdaudio.SoundId
	healAmount    float32
	ammoType      game.AmmoType
	ammoAmount    int
	dontWasteAmmo bool // If true, the item will not be collected if the player has maximum ammo
	onGround      bool
	fallSpeed     float32

	world *World
	id    scene.Id[*Item]
}

var _ comps.HasBody = (*Item)(nil)

func SpawnItemFromTE3(world *World, ent te3.Ent) (id scene.Id[*Item], item *Item, err error) {
	itemType, isItem := ent.Properties["item"]
	if !isItem {
		return scene.Id[*Item]{}, nil, fmt.Errorf("item is missing 'item' property")
	}

	switch itemType {
	case "medkit":
		id, item, err = SpawnMedkit(world, ent.Position)
	case "stimpack":
		id, item, err = SpawnStimpack(world, ent.Position)
	case "cartonofeggs", "egg_carton":
		id, item, err = SpawnEggCarton(world, ent.Position)
	default:
		return scene.Id[*Item]{}, nil, fmt.Errorf("item type '%v' is not implemented yet", itemType)
	}

	if err != nil {
		return
	}

	// Put the item on the floor using a raycast
	cast := world.GameMap.Body().Shape.Raycast(ent.Position, math2.Vec3Down(), world.GameMap.Body().Transform.Position(), 100.0)
	if cast.Hit {
		item.body.Transform.SetPosition(math2.Vec3WithY(cast.Position, cast.Position.Y()+item.body.Transform.Scale().Y()))
		item.onGround = true
	}

	return
}

func SpawnStimpack(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.25, 0.25, 0.25})
	item.healAmount = 10.0
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/stimpack.png"))
	return
}

func SpawnMedkit(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.375, 0.375, 0.375})
	item.healAmount = 50.0
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/medkit.png"))
	return
}

func SpawnEggCarton(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.5, 0.5, 0.5})
	item.ammoType = game.AMMO_TYPE_EGG
	item.ammoAmount = 12
	item.dontWasteAmmo = true
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/egg_carton.png"))
	return
}

func spawnItemGeneric(world *World, position, rotation, scale mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = world.Items.New()
	if err != nil {
		return
	}

	*item = Item{
		body: comps.Body{
			Transform:   comps.TransformFromTranslationAnglesScale(position, rotation, scale),
			Shape:       collision.NewSphere(0.5),
			Layer:       0,
			Filter:      0,
			OnIntersect: item.onIntersect,
		},
		flashColor:  color.White.WithAlpha(0.75),
		pickupSound: cache.GetSfx("assets/sounds/pickup.wav"),
		world:       world,
		id:          id,
		fallSpeed:   2.0,
	}

	return
}

func (item *Item) Body() *comps.Body {
	return &item.body
}

func (item *Item) Update(deltaTime float32) {
	item.animPlayer.Update(deltaTime)
	if !item.onGround {
		// Fall until the ground is touched.
		cast := item.world.GameMap.Body().Shape.Raycast(
			item.body.Transform.Position(),
			math2.Vec3Down(),
			item.world.GameMap.Body().Transform.Position(),
			item.body.Transform.Scale().Y()+(deltaTime*item.fallSpeed))
		if cast.Hit {
			item.body.Transform.SetPosition(math2.Vec3WithY(cast.Position, cast.Position.Y()+item.body.Transform.Scale().Y()))
			item.onGround = true
		} else {
			item.Body().Transform.Translate(0.0, -deltaTime*item.fallSpeed, 0.0)
		}
	}
}

func (item *Item) Render(context *render.Context) {
	item.spriteRender.Render(&item.body.Transform, &item.animPlayer, context, 0.0)
}

func (item *Item) onIntersect(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	player, isPlayer := otherEnt.(*Player)
	if !otherEnt.Body().OnLayer(COL_LAYER_PLAYERS) || !isPlayer {
		return
	}

	if item.ammoAmount != 0 && item.ammoType != game.AMMO_TYPE_NONE {
		notWasted := player.AddAmmo(item.ammoType, item.ammoAmount)
		if item.dontWasteAmmo && !notWasted {
			return
		}
	}
	player.actor.Health += item.healAmount

	item.pickupSound.PlayAttenuatedV(item.body.Transform.Position())
	item.world.Hud.FlashScreen(item.flashColor, 1.5)
	item.world.QueueRemoval(item.id.Handle)
}
