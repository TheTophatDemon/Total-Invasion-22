package game

/**============================================
 *               Keys
 *=============================================**/

type Keys uint8

const (
	KeysNone Keys = 0
	KeysBlue      = 1 << (iota - 1)
	KeysBrown
	KeysYellow
	KeysGray
	KeysCount
	KeysAll = KeysBlue | KeysBrown | KeysYellow | KeysGray
)

var keycardNames = [...]string{
	KeysBlue:   "blue",
	KeysBrown:  "brown",
	KeysYellow: "yellow",
	KeysGray:   "gray",
}

func (keys Keys) Name() string {
	for key, name := range keycardNames {
		if (int(keys) & key) != 0 {
			return name
		}
	}
	return "invalid"
}

func KeyTypeFromName(name string) Keys {
	for i, v := range keycardNames {
		if v == name {
			return Keys(i)
		}
	}
	return KeysNone
}

/**============================================
 *               Enemies
 *=============================================**/

type EnemyType uint8

const (
	EnemyTypeWraith EnemyType = iota
	EnemyTypeFireWraith
	EnemyTypeMotherWraith
	EnemyTypeDummkopf
	EnemyTypePrisrak
	EnemyTypeCount
)

func (et EnemyType) String() string {
	switch et {
	case EnemyTypeWraith:
		return "wraith"
	case EnemyTypeFireWraith:
		return "fire wraith"
	case EnemyTypeMotherWraith:
		return "mother wraith"
	case EnemyTypeDummkopf:
		return "dummkopf"
	case EnemyTypePrisrak:
		return "prisrak"
	default:
		return "unknown"
	}
}

/**============================================
 *               Armor
 *=============================================**/

const MaxArmorAmount = 200

type ArmorType struct {
	name            string
	defense         float32 // Fraction of damaged absorbed
	knockbackRange  float32 // Enemies will be damaged when touching you within this distance
	knockbackDamage float32 // Each time you knock back an enemy, it will be damaged this much
	knockbackForce  float32 // Multiplier for the force to knock back an enemy
	speedMultiplier float32 // Multiplier for the player's max speed while wearing the armor
}

func (armor ArmorType) Name() string {
	return armor.name
}

// Fraction of damaged absorbed
func (armor ArmorType) Defense() float32 {
	return armor.defense
}

// Enemies will be damaged when touching you within this distance
func (armor ArmorType) KnockbackRange() float32 {
	return armor.knockbackRange
}

// Each time you knock back an enemy, it will be damaged this much
func (armor ArmorType) KnockbackDamage() float32 {
	return armor.knockbackDamage
}

// Multiplier for the force to knock back an enemy
func (armor ArmorType) KnockbackForce() float32 {
	return armor.knockbackForce
}

// Multiplier for the player's max speed while wearing the armor
func (armor ArmorType) SpeedMultiplier() float32 {
	return armor.speedMultiplier
}

var (
	ArmorTypeNone   = ArmorType{}
	ArmorTypeBoring = ArmorType{
		name:    "boring",
		defense: 0.5,
	}
	ArmorTypeBullet = ArmorType{
		name:    "bullet",
		defense: 0.3,
	}
	ArmorTypeSuper = ArmorType{
		name:            "super",
		defense:         0.7,
		knockbackRange:  2,
		knockbackDamage: 20.0,
		knockbackForce:  7,
		speedMultiplier: 1.35,
	}
	ArmorTypeChronos = ArmorType{
		name:    "chronos",
		defense: 0.5,
	}
)

/**============================================
 *               Ammo
 *=============================================**/

type (
	AmmoType uint8
	Ammo     [AmmoTypeCount]int
)

const (
	AmmoTypeNone AmmoType = iota
	AmmoTypeSickle
	AmmoTypeEgg
	AmmoTypeGrenade
	AmmoTypePlasma
	AmmoTypeCount
)

// Returns the maximum amount of this ammo type that can be held.
func (ammoType AmmoType) Limit() int {
	switch ammoType {
	case AmmoTypeSickle:
		return 1
	case AmmoTypeEgg:
		return 255
	case AmmoTypeGrenade:
		return 100
	case AmmoTypePlasma:
		return 500
	}
	return 0
}

type (
	WeaponIndex int
	Equipment   struct {
		Ammo            Ammo
		Armor           ArmorType
		ArmorAmount     float32
		EquippedWeapons [WeaponIndexCount]bool
		Keys            Keys
		SelectedWeapon  WeaponIndex
	}
)

const (
	WeaponIndexNone WeaponIndex = iota
	WeaponIndexSickle
	WeaponIndexChicken
	WeaponIndexGrenade
	WeaponIndexParusu
	WeaponIndexDblGrenade
	WeaponIndexSign
	WeaponIndexAirhorn
	WeaponIndexDefenestrator
	WeaponIndexCluckster
	WeaponIndexCount
)
