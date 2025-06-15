package game

type ArmorType uint8

const (
	ARMOR_TYPE_NONE = iota
	ARMOR_TYPE_BORING
	ARMOR_TYPE_BULLET
	ARMOR_TYPE_SUPER
	ARMOR_TYPE_CHRONOS
)

const MAX_ARMOR = 200

var ArmorNames = [...]string{
	ARMOR_TYPE_BORING:  "boring",
	ARMOR_TYPE_BULLET:  "bullet",
	ARMOR_TYPE_SUPER:   "super",
	ARMOR_TYPE_CHRONOS: "chronos",
}

// Holds the fraction of damage absorbed for each armor type.
var ArmorDefense = [...]float32{
	ARMOR_TYPE_BORING:  0.5,
	ARMOR_TYPE_BULLET:  0.3,
	ARMOR_TYPE_SUPER:   0.75,
	ARMOR_TYPE_CHRONOS: 0.5,
}

func ArmorTypeFromName(name string) ArmorType {
	for i, v := range ArmorNames {
		if v == name {
			return ArmorType(i)
		}
	}
	return ARMOR_TYPE_NONE
}
