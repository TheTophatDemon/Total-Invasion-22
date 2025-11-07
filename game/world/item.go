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
	giveAmmo    [game.AmmoTypeCount]int // Amount of ammo to give for each type
	giveWeapon  game.WeaponType
	dontWaste   bool // If true, the item will not be collected if the player has the maximum of its resource
	giveKey     game.Keys
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

	id scene.Id[*Item]
}

var _ comps.HasBody = (*Item)(nil)

func SpawnItemFromTE3(ent te3.Ent) (id scene.Id[*Item], item *Item, err error) {
	itemType, isItem := ent.Properties["item"]
	if !isItem {
		return scene.Id[*Item]{}, nil, fmt.Errorf("item is missing 'item' property")
	}

	switch itemType {
	case "medkit":
		id, item, err = SpawnMedkit(ent.Position)
	case "stimpack":
		id, item, err = SpawnStimpack(ent.Position)
	case "cartonofeggs", "egg_carton":
		id, item, err = SpawnEggCarton(ent.Position)
	case "grenades":
		id, item, err = SpawnGrenades(ent.Position)
	case "plasmavial", "plasma_vial", "plasma vial":
		id, item, err = SpawnPlasmaVials(ent.Position)
	case "chickencannon", "chickengun", "chicken_cannon", "chicken_gun":
		id, item, err = SpawnChickenCannon(ent.Position)
	case "grenadelauncher", "grenade_launcher", "grenade launcher":
		id, item, err = SpawnGrenadeLauncher(ent.Position)
	case "parusu":
		id, item, err = SpawnParusu(ent.Position)
	case "airhorn":
		id, item, err = SpawnAirhorn(ent.Position)
	case "bluecard":
		id, item, err = SpawnKeycard(ent.Position, game.KeysBlue)
		return
	case "graycard":
		id, item, err = SpawnKeycard(ent.Position, game.KeysGray)
		return
	case "yellowcard":
		id, item, err = SpawnKeycard(ent.Position, game.KeysYellow)
		return
	case "browncard":
		id, item, err = SpawnKeycard(ent.Position, game.KeysBrown)
		return
	case "boringarmor", "boring armor", "boring_armor":
		id, item, err = SpawnArmorStand(ent.Position, game.ArmorTypeBoring)
		item.armorAmount = 100
		item.flashColor = color.FromBytes(170, 85, 0, 180)
		return
	case "bulletarmor", "bullet armor", "bullet_armor":
		id, item, err = SpawnArmorStand(ent.Position, game.ArmorTypeBullet)
		item.armorAmount = 120
		item.giveAmmo = [game.AmmoTypeCount]int{
			game.AmmoTypeEgg:     12,
			game.AmmoTypeGrenade: 5,
			game.AmmoTypePlasma:  30,
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
	cast := gWorld.GameMap.GridShape.Raycast(ent.Position, math2.Vec3Down(), 100.0, ColLayerMap)
	if cast.Hit {
		item.body.Position = math2.Vec3WithY(cast.Position, cast.Position.Y()+item.spriteRender.Scale()[1])
		item.onGround = true
	}

	return
}

func SpawnStimpack(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.25, 0.25, 0.25})
	item.healAmount = 10.0
	item.dontWaste = true
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/stimpack.png"), nil, &mgl32.Vec2{0.25, 0.25})
	return
}

func SpawnMedkit(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.375, 0.375, 0.375})
	item.healAmount = 50.0
	item.dontWaste = true
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/medkit.png"), nil, &mgl32.Vec2{0.375, 0.375})
	return
}

func SpawnAmmo(position mgl32.Vec3, ammoType game.AmmoType) (scene.Id[*Item], *Item, error) {
	switch ammoType {
	case game.AmmoTypeEgg:
		return SpawnEggCarton(position)
	case game.AmmoTypeGrenade:
		return SpawnGrenades(position)
	case game.AmmoTypePlasma:
		return SpawnPlasmaVials(position)
	}
	return scene.Id[*Item]{}, nil, nil
}

func SpawnEggCarton(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.5, 0.5, 0.5})
	item.giveAmmo[game.AmmoTypeEgg] = 12
	item.dontWaste = true
	item.message = settings.Localize("eggCartonGet")
	item.messageTime = 1.0
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/egg_carton.png"), nil, &mgl32.Vec2{0.5, 0.5})
	return
}

func SpawnGrenades(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.5, 0.5, 0.5})
	item.giveAmmo[game.AmmoTypeGrenade] = 3
	item.dontWaste = true
	item.message = settings.Localize("grenadesGet")
	item.messageTime = 1.0
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/grenades.png"), nil, &mgl32.Vec2{0.5, 0.5})
	return
}

func SpawnPlasmaVials(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.375, 0.25, 0.375})
	item.giveAmmo[game.AmmoTypePlasma] = 50
	item.dontWaste = true
	item.message = settings.Localize("plasmaVialsGet")
	item.messageTime = 1.0
	tex := cache.GetTexture("assets/textures/sprites/plasma_vials.png")
	item.spriteRender = comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.375, 0.25})
	item.animPlayer = comps.NewAnimationPlayer(tex.GetDefaultAnimation(), true)
	return
}

func SpawnChickenCannon(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.625, 0.25, 0.625})
	item.giveAmmo[game.AmmoTypeEgg] = 24
	item.giveWeapon = game.WeaponChicken
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("chickenCannonGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/chicken_cannon.png"), nil, &mgl32.Vec2{0.625, 0.25})
	return
}

func SpawnGrenadeLauncher(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.5, 0.25, 0.5})
	item.giveAmmo[game.AmmoTypeGrenade] = 5
	item.giveWeapon = game.WeaponGrenade
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("grenadeLauncherGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/grenade_launcher.png"), nil, &mgl32.Vec2{0.5, 0.25})
	return
}

func SpawnParusu(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.625, 0.25, 0.625})
	item.giveAmmo[game.AmmoTypePlasma] = 100
	item.giveWeapon = game.WeaponParusu
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("parusuGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/parusu.png"), nil, &mgl32.Vec2{0.625, 0.25})
	return
}

func SpawnAirhorn(position mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.5, 0.5, 0.5})
	item.giveWeapon = game.WeaponAirhorn
	item.pickupSound = cache.GetSfx("assets/sounds/weapon.wav")
	item.message = settings.Localize("airhornGet")
	item.messageTime = 1.5
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/airhorn.png"), nil, &mgl32.Vec2{0.5, 0.5})
	return
}

func SpawnKeycard(position mgl32.Vec3, keyType game.Keys) (id scene.Id[*Item], item *Item, err error) {
	if keyType == 0 {
		err = fmt.Errorf("no key type supplied")
		return
	}
	id, item, err = spawnItemGeneric(position, mgl32.Vec3{0.25, 0.25, 0.25})
	item.spriteRender = comps.NewSpriteRender(cache.GetTexture("assets/textures/sprites/"+keyType.Name()+"card.png"), nil, &mgl32.Vec2{0.25, 0.25})
	item.message = settings.Localize(keyType.Name() + "KeyGet")
	item.messageTime = 1.5
	item.giveKey = keyType
	item.fallSpeed = 0.0
	item.pickupSound = cache.GetSfx("assets/sounds/key.wav")
	item.floatSpeed = 2.0
	item.floatAmplitude = 0.15
	item.floatOrigin = position
	return
}

func SpawnArmorStand(position mgl32.Vec3, armorType game.ArmorType) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = gWorld.Items.New()
	if err != nil {
		return
	}
	tex := cache.GetTexture("assets/textures/sprites/armor_stand.png")
	anim, _ := tex.GetAnimation(armorType.Name() + "Armor")
	*item = Item{
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(0.6, 0.8, 0.6),
			Layers:   ColLayerMap | ColLayerUsable,
		},
		animPlayer:   comps.NewAnimationPlayer(anim, false),
		pickupSound:  cache.GetSfx("assets/sounds/armor.wav"),
		id:           id,
		fallSpeed:    2.0,
		message:      settings.Localize(armorType.Name() + "ArmorGet"),
		messageTime:  1.5,
		messageColor: color.Red,
		spriteRender: comps.NewSpriteRender(tex, nil, &mgl32.Vec2{0.4, 0.8}),
		giveArmor:    armorType,
		dontWaste:    true,
	}

	if rand.Float32() < 0.2 {
		item.collectAnim, _ = tex.GetAnimation("undress")
	} else {
		item.collectAnim, _ = tex.GetAnimation("collect")
	}

	return
}

func spawnItemGeneric(position, size mgl32.Vec3) (id scene.Id[*Item], item *Item, err error) {
	id, item, err = gWorld.Items.New()
	if err != nil {
		return
	}

	*item = Item{
		body: comps.Body{
			Position: position,
			Shape:    collision.NewBoxShape(size[0], size[1], size[2]),
			Layers:   ColLayerUsable,
		},
		flashColor:   color.White.WithAlpha(0.75),
		pickupSound:  cache.GetSfx("assets/sounds/pickup.wav"),
		id:           id,
		fallSpeed:    2.0,
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

		height := item.body.Shape.Extents().Size()[1] / 2.0

		// Fall until the ground is touched.
		cast := gWorld.GameMap.GridShape.Raycast(
			item.body.Position,
			math2.Vec3Down(),
			height+(deltaTime*item.fallSpeed),
			ColLayerMap,
		)
		if cast.Hit {
			item.body.Position = math2.Vec3WithY(cast.Position, cast.Position[1]+height)
			item.onGround = true
		} else {
			item.body.Translate(0.0, -deltaTime*item.fallSpeed, 0.0)
		}
	} else if item.floatAmplitude != 0.0 && item.floatSpeed != 0.0 {
		item.floatTimer += deltaTime
		item.body.Position = item.floatOrigin.Add(mgl32.Vec3{0.0, math2.Sin(item.floatTimer*item.floatSpeed) * item.floatAmplitude, 0.0})
	}
}

func (item *Item) Render(context *render.Context) {
	item.spriteRender.Render(item.body.Position, &item.animPlayer, context, 0.0)
}

func (item *Item) OnUse(player *Player) {
	if !item.collectAnim.IsNil() && item.animPlayer.IsPlayingAnim(item.collectAnim) {
		return
	}

	if item.giveWeapon != game.WeaponNone {
		weapon := gWorld.Hud.Weapons.Get(item.giveWeapon)
		if weapon != nil {
			weapon.Equipped = true
			gWorld.Hud.Weapons.Select(item.giveWeapon)
		}
	}

	wasted := false
	for ammoType, ammoAmount := range item.giveAmmo {
		if ammoType != int(game.AmmoTypeNone) && ammoAmount > 0 {
			succ := player.AddAmmo(game.AmmoType(ammoType), ammoAmount)
			wasted = wasted && !succ
		}
	}
	if item.dontWaste && wasted {
		return
	}

	if player.actor.Health < player.actor.TargetHealth {
		player.actor.Health += item.healAmount
	} else if item.healAmount > 0 && item.dontWaste {
		return
	}

	if item.giveKey != game.KeysNone {
		player.keys |= item.giveKey
	}

	if item.armorAmount != 0 && item.giveArmor != game.ArmorTypeNone {
		notWasted := player.AddArmor(item.giveArmor, item.armorAmount)
		if item.dontWaste && !notWasted {
			return
		}
	}

	item.pickupSound.PlayAttenuatedV(item.body.Position)
	gWorld.Hud.FlashScreen(item.flashColor, 1.5)
	if len(item.message) > 0 {
		gWorld.Hud.ShowMessage(item.message, item.messageTime, 10, item.messageColor)
	}

	if item.collectAnim.IsNil() {
		gWorld.QueueRemoval(item.id.Handle)
	} else {
		item.animPlayer.PlayNewAnim(item.collectAnim)
	}
}
