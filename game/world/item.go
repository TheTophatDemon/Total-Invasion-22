package world

import (
	"fmt"
	"math/rand"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/assets/te3"
	"tophatdemon.com/total-invasion-ii/engine/assets/textures"
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

	flashColor  color.Color
	pickupSound tdaudio.SoundId
	healAmount  float32
	giveAmmo    [game.AMMO_TYPE_COUNT]int // Amount of ammo to give for each type
	giveWeapon  hud.WeaponIndex
	dontWaste   bool // If true, the item will not be collected if the player has the maximum of its resource
	giveKey     game.KeyType
	giveArmor   game.ArmorType
	armorAmount int

	collectAnim textures.Animation // Animation that will player after the item is collected. It will prevent the item from being removed.

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
	case "grenades":
		id, item, err = SpawnGrenades(world, ent.Position)
	case "plasmavial", "plasma_vial", "plasma vial":
		id, item, err = SpawnPlasmaVials(world, ent.Position)
	case "chickencannon", "chickengun", "chicken_cannon", "chicken_gun":
		id, item, err = SpawnChickenCannon(world, ent.Position)
	case "grenadelauncher", "grenade_launcher", "grenade launcher":
		id, item, err = SpawnGrenadeLauncher(world, ent.Position)
	case "parusu":
		id, item, err = SpawnParusu(world, ent.Position)
	case "airhorn":
		id, item, err = SpawnAirhorn(world, ent.Position)
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
	case "boringarmor", "boring armor", "boring_armor":
		id, item, err = SpawnArmorStand(world, ent.Position, game.ARMOR_TYPE_BORING)
		item.armorAmount = 100
		item.flashColor = color.FromBytes(170, 85, 0, 180)
		return
	case "bulletarmor", "bullet armor", "bullet_armor":
		id, item, err = SpawnArmorStand(world, ent.Position, game.ARMOR_TYPE_BULLET)
		item.armorAmount = 120
		item.giveAmmo = [game.AMMO_TYPE_COUNT]int{
			game.AMMO_TYPE_EGG:     12,
			game.AMMO_TYPE_GRENADE: 5,
			game.AMMO_TYPE_PLASMA:  30,
		}
		item.flashColor = color.FromBytes(0, 113, 0, 180)
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

func SpawnAmmo(world *World, position mgl32.Vec3, ammoType game.AmmoType) (scene.Id[*Item], *Item, error) {
	switch ammoType {
	case game.AMMO_TYPE_EGG:
		return SpawnEggCarton(world, position)
	case game.AMMO_TYPE_GRENADE:
		return SpawnGrenades(world, position)
	case game.AMMO_TYPE_PLASMA:
		return SpawnPlasmaVials(world, position)
	}
	return scene.Id[*Item]{}, nil, nil
}

func SpawnEggCarton(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.5, 0.5, 0.5})
	item.giveAmmo[game.AMMO_TYPE_EGG] = 6
	item.dontWaste = true
	item.message = settings.Localize("eggCartonGet")
	item.messageTime = 1.0
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/egg_carton.png"))
	return
}

func SpawnGrenades(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.5, 0.5, 0.5})
	item.giveAmmo[game.AMMO_TYPE_GRENADE] = 3
	item.dontWaste = true
	item.message = settings.Localize("grenadesGet")
	item.messageTime = 1.0
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/grenades.png"))
	return
}

func SpawnPlasmaVials(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.375, 0.25, 0.5})
	item.giveAmmo[game.AMMO_TYPE_PLASMA] = 50
	item.dontWaste = true
	item.message = settings.Localize("plasmaVialsGet")
	item.messageTime = 1.0
	tex := cache.GetTexture("assets/textures/sprites/plasma_vials.png")
	item.spriteRender = comps.NewSpriteRender(tex)
	item.animPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	return
}

func SpawnChickenCannon(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.625, 0.25, 0.5})
	item.giveAmmo[game.AMMO_TYPE_EGG] = 24
	item.giveWeapon = hud.WEAPON_ORDER_CHICKEN
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("chickenCannonGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/chicken_cannon.png"))
	return
}

func SpawnGrenadeLauncher(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.5, 0.25, 0.5})
	item.giveAmmo[game.AMMO_TYPE_GRENADE] = 5
	item.giveWeapon = hud.WEAPON_ORDER_GRENADE
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("grenadeLauncherGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/grenade_launcher.png"))
	return
}

func SpawnParusu(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.625, 0.25, 0.5})
	item.giveAmmo[game.AMMO_TYPE_PLASMA] = 100
	item.giveWeapon = hud.WEAPON_ORDER_PARUSU
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("parusuGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/parusu.png"))
	return
}

func SpawnAirhorn(world *World, position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(world, position, mgl32.Vec3{}, mgl32.Vec3{0.5, 0.5, 0.5})
	item.giveWeapon = hud.WEAPON_ORDER_AIRHORN
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("airhornGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/airhorn.png"))
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

func SpawnArmorStand(world *World, position mgl32.Vec3, armorType game.ArmorType) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = world.Items.New()
	if err != nil {
		return
	}
	tex := cache.GetTexture("assets/textures/sprites/armor_stand.png")
	anim, _ := tex.GetAnimation(game.ArmorNames[armorType] + "Armor")
	*item = Item{
		body: comps.Body{
			Transform: comps.TransformFromTranslationAnglesScale(position, mgl32.Vec3{}, mgl32.Vec3{0.4, 0.8, 0.8}),
			Shape:     collision.NewCylinder(0.6, 1.0),
			Layer:     COL_LAYER_MAP,
		},
		animPlayer:   comps.NewAnimationPlayer(anim, false),
		pickupSound:  cache.GetSfx("assets/sounds/armor.wav"),
		world:        world,
		id:           id,
		fallSpeed:    2.0,
		message:      settings.Localize(game.ArmorNames[armorType] + "ArmorGet"),
		messageTime:  1.5,
		messageColor: color.Red,
		spriteRender: comps.NewSpriteRender(tex),
		giveArmor:    armorType,
		giveWeapon:   hud.WEAPON_ORDER_NONE,
		dontWaste:    true,
	}

	if rand.Float32() < 0.2 {
		item.collectAnim, _ = tex.GetAnimation("undress")
	} else {
		item.collectAnim, _ = tex.GetAnimation("collect")
	}

	switch armorType {
	case game.ARMOR_TYPE_BORING:
		item.armorAmount = 100
	case game.ARMOR_TYPE_BULLET:
		item.armorAmount = 120
	case game.ARMOR_TYPE_SUPER, game.ARMOR_TYPE_CHRONOS:
		item.armorAmount = 200
	}

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

func (item *Item) OnUse(player *Player) {
	if !item.collectAnim.IsNil() && item.animPlayer.IsPlayingAnim(item.collectAnim) {
		return
	}

	if item.giveWeapon != hud.WEAPON_ORDER_NONE {
		item.world.Hud.EquipWeapon(item.giveWeapon)
		item.world.Hud.SelectWeapon(item.giveWeapon)
	}

	wasted := false
	for ammoType, ammoAmount := range item.giveAmmo {
		if ammoType != int(game.AMMO_TYPE_NONE) && ammoAmount > 0 {
			succ := player.AddAmmo(game.AmmoType(ammoType), ammoAmount)
			wasted = wasted && !succ
		}
	}
	if item.dontWaste && wasted {
		return
	}

	player.actor.Health += item.healAmount

	if item.giveKey != game.KEY_TYPE_INVALID {
		player.keys |= item.giveKey
	}

	if item.armorAmount != 0 && item.giveArmor != game.ARMOR_TYPE_NONE {
		notWasted := player.AddArmor(item.giveArmor, item.armorAmount)
		if item.dontWaste && !notWasted {
			return
		}
	}

	item.pickupSound.PlayAttenuatedV(item.body.Transform.Position())
	item.world.Hud.FlashScreen(item.flashColor, 1.5)
	if len(item.message) > 0 {
		item.world.Hud.ShowMessage(item.message, item.messageTime, 10, item.messageColor)
	}

	if item.collectAnim.IsNil() {
		item.world.QueueRemoval(item.id.Handle)
	} else {
		item.animPlayer.PlayNewAnim(item.collectAnim)
	}
}

func (item *Item) onIntersect(otherEnt comps.HasBody, result collision.Result, deltaTime float32) {
	player, isPlayer := otherEnt.(*Player)
	if !otherEnt.Body().OnLayer(COL_LAYER_PLAYERS) || !isPlayer {
		return
	}

	item.OnUse(player)
}
