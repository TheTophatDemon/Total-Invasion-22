package game

type KeyType uint8

const (
	KEY_TYPE_INVALID KeyType = 0
	KEY_TYPE_BLUE            = 1 << (iota - 1)
	KEY_TYPE_BROWN
	KEY_TYPE_YELLOW
	KEY_TYPE_GRAY
	KEY_TYPE_ALL = KEY_TYPE_BLUE | KEY_TYPE_BROWN | KEY_TYPE_YELLOW | KEY_TYPE_GRAY
)

var KeycardNames = [...]string{
	KEY_TYPE_BLUE:   "blue",
	KEY_TYPE_BROWN:  "brown",
	KEY_TYPE_YELLOW: "yellow",
	KEY_TYPE_GRAY:   "gray",
}

func KeyTypeFromName(name string) KeyType {
	for i, v := range KeycardNames {
		if v == name {
			return KeyType(i)
		}
	}
	return KEY_TYPE_INVALID
}
