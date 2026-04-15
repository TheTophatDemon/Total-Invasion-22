package hud

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"
	"tophatdemon.com/total-invasion-ii/engine/assets/cache"
	"tophatdemon.com/total-invasion-ii/engine/color"
	"tophatdemon.com/total-invasion-ii/engine/math2"
	"tophatdemon.com/total-invasion-ii/engine/scene/comps/ui/v2"
	"tophatdemon.com/total-invasion-ii/engine/tdaudio"
	"tophatdemon.com/total-invasion-ii/game"
)

type weaponState uint8

const (
	weaponStateInactive weaponState = iota
	weaponStateIntro
	weaponStateReady
	weaponStateOutro
	weaponStateCount
)

const (
	idleAnimName = "idle"
	fireAnimName = "fire"
)

type Weapon struct {
	Equipped        bool
	kind            game.WeaponType
	name            string // Weapon name, same as localization key.
	texturePath     string
	initialAnimName string // Name of the animation played after the weapon is initialized. If unset, will be "idle"
	sprite          ui.Element
	cooldown        float32
	cooldownTimer   float32
	sway            float32     // Value tracking the timeline of the sway animation.
	swayExtents     mgl32.Vec2  // Defines a rectangle on screen within which the weapon will sway
	swaySpeed       mgl32.Vec2  // Defines the speed at which the weapon will sway in each axis
	spriteOffset    mgl32.Vec2  // Offset from the bottom center of the screen where the sprite will rest
	state           weaponState // Describes the state of transitional animations.
	ammoType        game.AmmoType
	ammoCost        int         // Amount subtracted from ammo after firing
	wheelColor      color.Color // Color of weapon wheel frame
	wheelIconPath   string      // Path to icon texture displayed in weapon wheel slot
	noiseLevel      float32
	isShooter       bool
	fireSoundPath   string
	voice           tdaudio.VoiceId
	heldDown        bool // True while the fire button is being held
	updateFunc      func(w *Weapon, deltaTime float32, ammo game.Ammo)
}

func (weap *Weapon) Kind() game.WeaponType {
	return weap.kind
}

func (weap *Weapon) IsShooter() bool {
	return weap.isShooter
}

func (weap *Weapon) NoiseLevel() float32 {
	return weap.noiseLevel
}

func (weap *Weapon) init() {
	if len(weap.texturePath) == 0 {
		return
	}
	texture := cache.GetTexture(weap.texturePath)
	if len(weap.initialAnimName) == 0 {
		weap.initialAnimName = idleAnimName
	}
	initAnim, ok := texture.GetAnimation(weap.initialAnimName)
	if !ok {
		log.Printf("weapon %v missing idle animation\n", weap.name)
	}

	weap.sprite = ui.NewBox(ui.Transform{
		Anchor: ui.Ratios{0.5, 1.0},
		Origin: ui.Ratios{0.5, 1.0},
		Size: mgl32.Vec2{
			initAnim.Frames[0].Rect.Width * SpriteScale(),
			initAnim.Frames[0].Rect.Height * SpriteScale(),
		},
	}, texture)
	weap.sprite.SetPosition(weap.startPos())
	weap.sprite.AnimPlayer.PlayNewAnim(initAnim)
}

// The position that the weapon sprite will head towards after being selected
func (weap *Weapon) endPos() mgl32.Vec2 {
	return weap.spriteOffset
}

// The position that the weapon sprite will head towards after being deselected
func (weap *Weapon) startPos() mgl32.Vec2 {
	return weap.endPos().Add(mgl32.Vec2{0.0, weap.sprite.Height()})
}

func (weap *Weapon) onSelect() {
	if len(weap.texturePath) == 0 {
		return
	}
	weap.sway = 0.0
	weap.state = weaponStateIntro

	if weap.sprite.AnimPlayer.CurrentAnimation().Name != weap.initialAnimName {
		idleAnim, _ := weap.sprite.BgTexture.GetAnimation(idleAnimName)
		weap.sprite.AnimPlayer.PlayNewAnim(idleAnim)
	}
}

func (weap *Weapon) onDeselect() {
	if weap.state != weaponStateInactive {
		weap.state = weaponStateOutro
	}
}

func (weap *Weapon) isSelected() bool {
	return weap.state != weaponStateInactive
}

func (weap *Weapon) update(deltaTime float32, swayAmount float32, ammo game.Ammo) {
	if len(weap.texturePath) == 0 {
		return
	}
	weap.sprite.AnimPlayer.Update(deltaTime)
	weap.cooldownTimer = max(weap.cooldownTimer-deltaTime, 0.0)
	swayOfs := mgl32.Vec2{
		math2.Cos(weap.sway*weap.swaySpeed[0]) * weap.swayExtents[0],
		math2.Sin(weap.sway*weap.swaySpeed[1]) * weap.swayExtents[1],
	}
	endPos := weap.endPos().Add(swayOfs)
	startPos := weap.startPos().Add(swayOfs)
	switch weap.state {
	case weaponStateReady:
		weap.sway += deltaTime * swayAmount
		// Sway the weapon according to player movement
		weap.sprite.SetPosition(endPos)
	case weaponStateIntro, weaponStateOutro:
		// Move the weapon towards its screen position.
		target := endPos
		if weap.state == weaponStateOutro {
			target = startPos
		}
		diff := target.Sub(weap.sprite.Position())
		dist := diff.Len()
		moveAmt := deltaTime * 3072.0
		if dist < moveAmt {
			weap.sprite.SetPosition(target)
			weap.state = (weap.state + 1) % weaponStateCount
		} else {
			weap.sprite.Translate(diff.Mul(moveAmt / dist))
		}
	}

	if weap.updateFunc != nil {
		weap.updateFunc(weap, deltaTime, ammo)
	} else {
		// Default behavior: Swtitch to idle animation when firing is complete.
		idleAnim, _ := weap.sprite.BgTexture.GetAnimation(idleAnimName)
		if weap.canFire(ammo) || ammo[weap.ammoType] == 0 {
			weap.sprite.AnimPlayer.ChangeAnimation(idleAnim)
		}
	}

	weap.heldDown = false
}

func (weap *Weapon) fire(ammo *game.Ammo) {
	if len(weap.texturePath) == 0 {
		return
	}
	weap.cooldownTimer = weap.cooldown
	ammo[weap.ammoType] -= weap.ammoCost
	weap.heldDown = true

	fireAnim, ok := weap.sprite.BgTexture.GetAnimation(fireAnimName)
	if !ok {
		log.Printf("fire anim not found for weapon %v\n", weap.name)
		return
	}

	if weap.sprite.AnimPlayer.CurrentAnimation().Name != fireAnim.Name {
		weap.sprite.AnimPlayer.PlayNewAnim(fireAnim)
		if len(weap.fireSoundPath) > 0 {
			weap.voice = cache.GetSfx(weap.fireSoundPath).Play()
		}
	}
}

func (weap *Weapon) canFire(ammo game.Ammo) bool {
	return weap.state == weaponStateReady && weap.cooldownTimer <= 0.0 && ammo[weap.ammoType] >= weap.ammoCost
}
