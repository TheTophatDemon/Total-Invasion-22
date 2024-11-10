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
	"tophatdemon.com/total-invasion-ii/game/hud"
	"tophatdemon.com/total-invasion-ii/game/settings"
)

// Represents any object that can be 'picked up', like health and weapons.
type Item struct {
	body         comps.Body
	spriteRender comps.SpriteRender
	animPlayer   comps.AnimationPlayer

	flashColor    color.Color
	pickupSound   tdaudio.SoundId
	healAmount    float32
	ammoType      game.AmmoType
	ammoAmount    int
	giveWeapon    hud.WeaponIndex
	dontWasteAmmo bool // If true, the item will not be collected if the player has maximum ammo
	giveKey       game.KeyType

	onGround                               bool
	fallSpeed                              float32
	floatSpeed, floatAmplitude, floatTimer float32
	floatOrigin                            mgl32.Vec3

	message      string
	messageTime  float32
	messageColor color.Color

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
	case "chickencannon", "chickengun", "chicken_cannon", "chicken_gun":
		id, item, err = SpawnChickenCannon(world, ent.Position)
	case "grenadelauncher", "grenade_launcher", "grenade launcher":
		id, item, err = SpawnGrenadeLauncher(world, ent.Position)
	case "bluecard":
		id, item, err = SpawnKeycard(world, ent.Position, game.KEY_TYPE_BLUE)
		return
	case "graycard":
		id, item, err = SpawnKeycard(world, ent.Position, game.KEY_TYPE_GRAY)
		return
	case "yellowcard":
		id, item, err = SpawnKeycard(world, ent.Position, game.KEY_TYPE_YELLOW)
		return
	case "browncard":
		id, item, err = SpawnKeycard(world, ent.Position, game.KEY_TYPE_BROWN)
		return
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
	item.message = settings.Localize("eggCartonGet")
	item.messageTime = 1.0
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/egg_carton.png"))
	return
}

func SpawnChickenCannon(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.625, 0.25, 0.5})
	item.ammoType = game.AMMO_TYPE_EGG
	item.ammoAmount = 48
	item.giveWeapon = hud.WEAPON_ORDER_CHICKEN
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("chickenCannonGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/chicken_cannon.png"))
	return
}

func SpawnGrenadeLauncher(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.5, 0.25, 0.5})
	item.ammoType = game.AMMO_TYPE_GRENADE
	item.ammoAmount = 5
	item.giveWeapon = hud.WEAPON_ORDER_GRENADE
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("grenadeLauncherGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/grenade_launcher.png"))
	return
}

func SpawnKeycard(world *World, position mgl32.Vec3, keyType game.KeyType) (id scene.Id[*Item], item *Item, err error) {
	if keyType == 0 {
		err = fmt.Errorf("no key type supplied")
		return
	}
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.25, 0.25, 0.25})
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/" + game.KeycardNames[keyType] + "card.png"))
	item.message = settings.Localize(game.KeycardNames[keyType] + "KeyGet")
	item.messageTime = 1.5
	item.giveKey = keyType
	item.fallSpeed = 0.0
	item.pickupSound = cache.GetSfx("assets/sounds/key.wav")
	item.floatSpeed = 2.0
	item.floatAmplitude = 0.15
	item.floatOrigin = position
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
		flashColor:   color.White.WithAlpha(0.75),
		pickupSound:  cache.GetSfx("assets/sounds/pickup.wav"),
		world:        world,
		id:           id,
		fallSpeed:    2.0,
		giveWeapon:   hud.WEAPON_ORDER_NONE,
		messageColor: color.Red,
	}

	return
}

func (item *Item) Body() *comps.Body {
	return &item.body
}

func (item *Item) Update(deltaTime float32) {
	item.animPlayer.Update(deltaTime)
	if !item.onGround && item.fallSpeed != 0.0 {
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
	} else if item.floatAmplitude != 0.0 && item.floatSpeed != 0.0 {
		item.floatTimer += deltaTime
		item.body.Transform.SetPosition(item.floatOrigin.Add(mgl32.Vec3{0.0, math2.Sin(item.floatTimer*item.floatSpeed) * item.floatAmplitude, 0.0}))
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

	if item.giveWeapon != hud.WEAPON_ORDER_NONE {
		item.world.Hud.EquipWeapon(item.giveWeapon)
		item.world.Hud.SelectWeapon(item.giveWeapon)
	}
	if item.ammoAmount != 0 && item.ammoType != game.AMMO_TYPE_NONE {
		notWasted := player.AddAmmo(item.ammoType, item.ammoAmount)
		if item.dontWasteAmmo && !notWasted {
			return
		}
	}
	player.actor.Health += item.healAmount

	if item.giveKey != game.KEY_TYPE_INVALID {
		player.keys |= item.giveKey
	}

	item.pickupSound.PlayAttenuatedV(item.body.Transform.Position())
	item.world.Hud.FlashScreen(item.flashColor, 1.5)
	if len(item.message) > 0 {
		item.world.Hud.ShowMessage(item.message, item.messageTime, 10, item.messageColor)
	}

	item.world.QueueRemoval(item.id.Handle)
}
