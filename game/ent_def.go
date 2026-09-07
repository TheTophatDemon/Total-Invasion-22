package game

import "tophatdemon.com/total-invasion-ii/engine/assets/te3"

// Defines properties available to a .te3 map's entity definitions
type EntProps struct {
	// Player
	AirhornEquipped       te3.BoolProp   `json:"airhorn equipped"`
	AmmoEgg               te3.IntProp    `json:"ammo egg"`
	AmmoGrenade           te3.IntProp    `json:"ammo grenade"`
	AmmoPlasma            te3.IntProp    `json:"ammo plasma"`
	Armor                 te3.StringProp `json:"armor"`
	ArmorAmount           te3.FloatProp  `json:"armor amount"`
	ChickenEquipped       te3.BoolProp   `json:"chicken equipped"`
	ClucksterEquipped     te3.BoolProp   `json:"cluckster equipped"`
	DefenestratorEquipped te3.BoolProp   `json:"defenestrator equipped"`
	DoubleGrenadeEquipped te3.BoolProp   `json:"double grenade equipped"`
	GrenadeEquipped       te3.BoolProp   `json:"grenade equipped"`
	Keys                  te3.IntProp    `json:"keys"`
	ParusuEquipped        te3.BoolProp   `json:"parusu equipped"`
	SignEquipped          te3.BoolProp   `json:"sign equipped"`
	// Level properties
	Sky  te3.StringProp `json:"sky"`
	Song te3.StringProp `json:"song"`
	// Walls and doors
	ActivateSound te3.StringProp `json:"activate sound"`
	Direction     te3.StringProp `json:"direction"`
	Distance      te3.FloatProp  `json:"distance"`
	Key           te3.StringProp `json:"key"`
	Open          te3.BoolProp   `json:"open"`
	On            te3.BoolProp   `json:"on"`
	Speed         te3.FloatProp  `json:"speed"`
	Unopenable    te3.BoolProp   `json:"unopenable"`
	// Triggers
	Action          te3.StringProp `json:"action"`
	DamagePerSecond te3.FloatProp  `json:"damage per second"`
	Level           te3.StringProp `json:"level"`
	// Props
	Prop   te3.StringProp `json:"prop"`
	Radius te3.FloatProp  `json:"radius"`
	// Items
	CollectAnim      te3.StringProp `json:"collect anim"`
	CollectAnimFrame te3.IntProp    `json:"collect anim frame"`
	Item             te3.StringProp `json:"item"`
	// Enemy
	Enemy te3.StringProp `json:"enemy"`
	// Generic
	Health          te3.FloatProp  `json:"health"`
	Link            te3.IntProp    `json:"link"`
	MaxDifficulty   te3.IntProp    `json:"max difficulty"`
	MinDifficulty   te3.IntProp    `json:"min difficulty"`
	MessageColor    te3.Vec3Prop   `json:"message color"`
	MessageKey      te3.StringProp `json:"message key"`
	MessagePriority te3.IntProp    `json:"message priority"`
	Name            te3.StringProp `json:"name"`
	Texture         te3.StringProp `json:"texture"`
	Type            te3.StringProp `json:"type"`
	Wait            te3.FloatProp  `json:"wait"`
}

// Defines a type for the .te3 map file's entity definitions.
type EntDef struct {
	te3.Ent[EntProps]
}
